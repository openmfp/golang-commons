package lifecycle

import (
	"context"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/api/equality"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/openmfp/golang-commons/controller/filter"
	"github.com/openmfp/golang-commons/errors"
	"github.com/openmfp/golang-commons/logger"
	"github.com/openmfp/golang-commons/sentry"
)

type LifecycleManager struct {
	log                *logger.Logger
	client             client.Client
	subroutines        []Subroutine
	operatorName       string
	controllerName     string
	spreadReconciles   bool
	manageConditions   bool
	readOnly           bool
	prepareContextFunc PrepareContextFunc
}

type RuntimeObject interface {
	runtime.Object
	v1.Object
}

type Subroutine interface {
	Process(ctx context.Context, instance RuntimeObject) (ctrl.Result, errors.OperatorError)
	Finalize(ctx context.Context, instance RuntimeObject) (ctrl.Result, errors.OperatorError)
	GetName() string
	Finalizers() []string
}

func NewLifecycleManager(log *logger.Logger, operatorName string, controllerName string, client client.Client, subroutines []Subroutine) *LifecycleManager {

	log = log.MustChildLoggerWithAttributes("operator", operatorName, "controller", controllerName)
	return &LifecycleManager{
		log:              log,
		client:           client,
		subroutines:      subroutines,
		operatorName:     operatorName,
		controllerName:   controllerName,
		spreadReconciles: false,
	}
}

func (l *LifecycleManager) Reconcile(ctx context.Context, req ctrl.Request, instance RuntimeObject) (ctrl.Result, error) {
	ctx, span := otel.Tracer(l.operatorName).Start(ctx, fmt.Sprintf("%s.Reconcile", l.controllerName))
	defer span.End()

	result := ctrl.Result{}
	reconcileId := uuid.New().String()

	log := l.log.MustChildLoggerWithAttributes("name", req.Name, "namespace", req.Namespace, "reconcile_id", reconcileId)
	sentryTags := sentry.Tags{"namespace": req.Namespace, "name": req.Name}

	ctx = logger.SetLoggerInContext(ctx, log)
	ctx = sentry.ContextWithSentryTags(ctx, sentryTags)

	log.Info().Msg("start reconcile")
	generationChanged := true
	err := l.client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if kerrors.IsNotFound(err) {
			log.Info().Msg("instance not found. It was likely deleted")
			return ctrl.Result{}, nil
		}
		return l.handleClientError("failed to retrieve instance", log, err, generationChanged, sentryTags)
	}

	originalCopy := instance.DeepCopyObject()
	inDeletion := instance.GetDeletionTimestamp() != nil

	if l.spreadReconciles && instance.GetDeletionTimestamp().IsZero() {
		instanceStatusObj := MustToRuntimeObjectSpreadReconcileStatusInterface(instance, log)
		generationChanged = instance.GetGeneration() != instanceStatusObj.GetObservedGeneration()
		isAfterNextReconcileTime := v1.Now().UTC().After(instanceStatusObj.GetNextReconcileTime().Time.UTC())
		refreshRequested := slices.Contains(maps.Keys(instance.GetLabels()), SpreadReconcileRefreshLabel)

		reconcileRequired := generationChanged || isAfterNextReconcileTime || refreshRequested
		if !reconcileRequired {
			log.Info().Msg("skipping reconciliation, spread reconcile is active. No processing needed")
			return onNextReconcile(instanceStatusObj, log)
		}
	}

	// Manage Finalizers
	ferr := l.addFinalizersIfNeeded(ctx, instance)
	if ferr != nil {
		return ctrl.Result{}, ferr
	}

	var conditions []v1.Condition
	if l.manageConditions {
		conditions = MustToRuntimeObjectConditionsInterface(instance, log).GetConditions()
		setInstanceConditionUnknownIfNotSet(&conditions)
	}

	if l.prepareContextFunc != nil {
		localCtx, oErr := l.prepareContextFunc(ctx, instance)
		if oErr != nil {
			return l.handleOperatorError(ctx, oErr, "failed to prepare context", generationChanged)
		}
		ctx = localCtx
	}

	// In case of deletion execute the finalize subroutines in the reverse order as subroutine processing
	subroutines := make([]Subroutine, len(l.subroutines))
	copy(subroutines, l.subroutines)
	if inDeletion {
		slices.Reverse(subroutines)
	}

	// Continue with reconciliation
	for _, subroutine := range subroutines {
		if l.manageConditions {
			setSubroutineConditionToUnknownIfNotSet(&conditions, subroutine, inDeletion, log)
		}

		// Set current conditions before reconciling the subroutine
		if l.manageConditions {
			MustToRuntimeObjectConditionsInterface(instance, log).SetConditions(conditions)
		}
		subResult, retry, err := l.reconcileSubroutine(ctx, instance, subroutine, log, generationChanged, sentryTags)
		// Update conditions with any changes the subroutine did
		if l.manageConditions {
			conditions = MustToRuntimeObjectConditionsInterface(instance, log).GetConditions()
		}
		if err != nil {
			if l.manageConditions {
				setSubroutineCondition(&conditions, subroutine, result, err, inDeletion, log)
				setInstanceConditionReady(&conditions, v1.ConditionFalse)
				MustToRuntimeObjectConditionsInterface(instance, log).SetConditions(conditions)
			}
			if !retry {
				l.markResourceAsFinal(instance, log, conditions, v1.ConditionFalse)
			}
			if !l.readOnly {
				_ = updateStatus(ctx, l.client, originalCopy, instance, log, generationChanged, sentryTags)
			}
			if !retry {
				return ctrl.Result{}, nil
			}
			return subResult, err
		}
		if subResult.Requeue {
			result.Requeue = subResult.Requeue
		}
		if subResult.RequeueAfter > 0 {
			if subResult.RequeueAfter < result.RequeueAfter || result.RequeueAfter == 0 {
				result.RequeueAfter = subResult.RequeueAfter
			}
		}
		if l.manageConditions {
			if !subResult.Requeue && subResult.RequeueAfter == 0 {
				setSubroutineCondition(&conditions, subroutine, subResult, err, inDeletion, log)
			}
		}
	}

	if !result.Requeue && result.RequeueAfter == 0 {
		// Reconciliation was successful
		l.markResourceAsFinal(instance, log, conditions, v1.ConditionTrue)
	} else {
		if l.manageConditions {
			setInstanceConditionReady(&conditions, v1.ConditionFalse)
		}
	}

	if l.manageConditions {
		MustToRuntimeObjectConditionsInterface(instance, log).SetConditions(conditions)
	}

	if !l.readOnly {
		err = updateStatus(ctx, l.client, originalCopy, instance, log, generationChanged, sentryTags)
		if err != nil {
			return result, err
		}
	}

	if l.spreadReconciles && instance.GetDeletionTimestamp().IsZero() {
		removed := removeRefreshLabelIfExists(instance)
		if removed {
			updateErr := l.client.Update(ctx, instance)
			if updateErr != nil {
				return l.handleClientError("failed to update instance", log, err, generationChanged, sentryTags)
			}
		}
	}

	log.Info().Msg("end reconcile")
	return result, nil
}

func (l *LifecycleManager) markResourceAsFinal(instance RuntimeObject, log *logger.Logger, conditions []v1.Condition, status v1.ConditionStatus) {
	if l.spreadReconciles && instance.GetDeletionTimestamp().IsZero() {
		instanceStatusObj := MustToRuntimeObjectSpreadReconcileStatusInterface(instance, log)
		setNextReconcileTime(instanceStatusObj, log)
		updateObservedGeneration(instanceStatusObj, log)
	}

	if l.manageConditions {
		setInstanceConditionReady(&conditions, status)
	}
}

func (l *LifecycleManager) validateInterfaces(instance RuntimeObject, log *logger.Logger) error {
	if l.spreadReconciles {
		_, err := toRuntimeObjectSpreadReconcileStatusInterface(instance, log)
		if err != nil {
			return err
		}
	}
	if l.manageConditions {
		_, err := toRuntimeObjectConditionsInterface(instance, log)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateStatus(ctx context.Context, cl client.Client, original runtime.Object, current RuntimeObject, log *logger.Logger, generationChanged bool, sentryTags sentry.Tags) error {
	currentUn, err := runtime.DefaultUnstructuredConverter.ToUnstructured(current)
	if err != nil {
		return err
	}

	originalUn, err := runtime.DefaultUnstructuredConverter.ToUnstructured(original)
	if err != nil {
		return err
	}

	currentStatus, hasField, err := unstructured.NestedFieldCopy(currentUn, "status")
	if err != nil {
		return err
	}
	if !hasField {
		return fmt.Errorf("status field not found in current object")
	}

	originalStatus, hasField, err := unstructured.NestedFieldCopy(originalUn, "status")
	if err != nil {
		return err
	}
	if !hasField {
		return fmt.Errorf("status field not found in current object")
	}

	if equality.Semantic.DeepEqual(currentStatus, originalStatus) {
		log.Info().Msg("skipping status update, since they are equal")
		return nil
	}

	log.Info().Msg("updating resource status")
	err = cl.Status().Update(ctx, current)
	if err != nil {
		if !kerrors.IsConflict(err) {
			log.Error().Err(err).Msg("cannot update status, kubernetes client error")
			if generationChanged {
				sentry.CaptureError(err, sentryTags, sentry.Extras{"message": "Updating of instance status failed"})
			}
		}
		log.Error().Err(err).Msg("cannot update reconciliation Conditions, kubernetes client error")
		return err
	}

	return nil
}

func (l *LifecycleManager) handleOperatorError(ctx context.Context, operatorError errors.OperatorError, msg string, generationChanged bool) (ctrl.Result, error) {
	l.log.Error().Bool("retry", operatorError.Retry()).Bool("sentry", operatorError.Sentry()).Err(operatorError.Err()).Msg(msg)
	if generationChanged && operatorError.Sentry() {
		sentry.CaptureError(operatorError.Err(), sentry.GetSentryTagsFromContext(ctx))
	}

	if operatorError.Retry() {
		return ctrl.Result{}, operatorError.Err()
	}

	return ctrl.Result{}, nil
}

func (l *LifecycleManager) handleClientError(msg string, log *logger.Logger, err error, generationChanged bool, sentryTags sentry.Tags) (ctrl.Result, error) {
	log.Error().Err(err).Msg(msg)
	if generationChanged {
		sentry.CaptureError(err, sentryTags)
	}

	return ctrl.Result{}, err
}

func containsFinalizer(o client.Object, subroutineFinalizers []string) bool {
	for _, subroutineFinalizer := range subroutineFinalizers {
		if controllerutil.ContainsFinalizer(o, subroutineFinalizer) {
			return true
		}
	}
	return false
}

func (l *LifecycleManager) reconcileSubroutine(ctx context.Context, instance RuntimeObject, subroutine Subroutine, log *logger.Logger, generationChanged bool, sentryTags map[string]string) (ctrl.Result, bool, error) {
	subroutineLogger := log.ChildLogger("subroutine", subroutine.GetName())
	ctx = logger.SetLoggerInContext(ctx, subroutineLogger)
	subroutineLogger.Debug().Msg("start subroutine")

	ctx, span := otel.Tracer(l.operatorName).Start(ctx, fmt.Sprintf("%s.reconcileSubroutine.%s", l.controllerName, subroutine.GetName()))
	defer span.End()
	var result ctrl.Result
	var err errors.OperatorError
	if instance.GetDeletionTimestamp() != nil {
		if containsFinalizer(instance, subroutine.Finalizers()) {
			subroutineLogger.Debug().Msg("finalizing instance")
			result, err = subroutine.Finalize(ctx, instance)
			subroutineLogger.Debug().Any("result", result).Msg("finalized instance")
			if err == nil {
				// Remove finalizers unless requeue is requested
				err = l.removeFinalizerIfNeeded(ctx, instance, subroutine, result)
			}
		}
	} else {
		subroutineLogger.Debug().Msg("processing instance")
		result, err = subroutine.Process(ctx, instance)
		subroutineLogger.Debug().Any("result", result).Msg("processed instance")
	}

	if err != nil {
		if generationChanged && err.Sentry() {
			sentry.CaptureError(err.Err(), sentryTags)
		}
		subroutineLogger.Error().Err(err.Err()).Bool("retry", err.Retry()).Msg("subroutine ended with error")
		return result, err.Retry(), err.Err()
	}

	subroutineLogger.Debug().Msg("end subroutine")
	return result, false, nil
}

func (l *LifecycleManager) addFinalizersIfNeeded(ctx context.Context, instance RuntimeObject) error {
	if l.readOnly {
		return nil
	}

	if !instance.GetDeletionTimestamp().IsZero() {
		return nil
	}

	update := false
	for _, subroutine := range l.subroutines {
		if len(subroutine.Finalizers()) > 0 {
			needsUpdate := l.addFinalizerIfNeeded(instance, subroutine)
			if needsUpdate {
				update = true
			}
		}
	}
	if update {
		err := l.client.Update(ctx, instance)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *LifecycleManager) addFinalizerIfNeeded(instance RuntimeObject, subroutine Subroutine) bool {
	update := false
	for _, f := range subroutine.Finalizers() {
		needsUpdate := controllerutil.AddFinalizer(instance, f)
		if needsUpdate {
			update = true
		}
	}
	return update
}

func (l *LifecycleManager) removeFinalizerIfNeeded(ctx context.Context, instance RuntimeObject, subroutine Subroutine, result ctrl.Result) errors.OperatorError {
	if l.readOnly {
		return nil
	}

	if !result.Requeue && result.RequeueAfter == 0 {
		update := false
		for _, f := range subroutine.Finalizers() {
			needsUpdate := controllerutil.RemoveFinalizer(instance, f)
			if needsUpdate {
				update = true
			}
		}
		if update {
			err := l.client.Update(ctx, instance)
			if err != nil {
				return errors.NewOperatorError(errors.Wrap(err, "failed to update instance"), true, false)
			}
		}
	}

	return nil
}

func (l *LifecycleManager) SetupWithManagerBuilder(mgr ctrl.Manager, maxReconciles int, reconcilerName string, instance RuntimeObject, debugLabelValue string, log *logger.Logger, eventPredicates ...predicate.Predicate) (*builder.Builder, error) {
	if err := l.validateInterfaces(instance, log); err != nil {
		return nil, err
	}

	if (l.manageConditions || l.spreadReconciles) && l.readOnly {
		return nil, fmt.Errorf("cannot use conditions or spread reconciles in read-only mode")
	}

	eventPredicates = append([]predicate.Predicate{filter.DebugResourcesBehaviourPredicate(debugLabelValue)}, eventPredicates...)
	return ctrl.NewControllerManagedBy(mgr).
		Named(reconcilerName).
		For(instance).
		WithOptions(controller.Options{MaxConcurrentReconciles: maxReconciles}).
		WithEventFilter(predicate.And(eventPredicates...)), nil
}

func (l *LifecycleManager) SetupWithManager(mgr ctrl.Manager, maxReconciles int, reconcilerName string, instance RuntimeObject, debugLabelValue string, r reconcile.Reconciler, log *logger.Logger, eventPredicates ...predicate.Predicate) error {
	bldr, err := l.SetupWithManagerBuilder(mgr, maxReconciles, reconcilerName, instance, debugLabelValue, log, eventPredicates...)
	if err != nil {
		return err
	}

	return bldr.Complete(r)
}

type PrepareContextFunc func(ctx context.Context, instance RuntimeObject) (context.Context, errors.OperatorError)

// WithPrepareContextFunc allows to set a function that prepares the context before each reconciliation
// This can be used to add additional information to the context that is needed by the subroutines
// You need to return a new context and an OperatorError in case of an error
func (l *LifecycleManager) WithPrepareContextFunc(prepareFunction PrepareContextFunc) *LifecycleManager {
	l.prepareContextFunc = prepareFunction
	return l
}

// WithReadOnly allows to set the controller to read-only mode
// In read-only mode, the controller will not update the status of the instance
func (l *LifecycleManager) WithReadOnly() *LifecycleManager {
	l.readOnly = true
	return l
}
