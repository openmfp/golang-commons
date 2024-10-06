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
	err := l.client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if kerrors.IsNotFound(err) {
			log.Info().Msg("instance not found. It was likely deleted")
			return ctrl.Result{}, nil
		}
		return l.handleClientError("failed to retrieve instance", log, err, sentryTags)
	}

	originalCopy := instance.DeepCopyObject()
	inDeletion := instance.GetDeletionTimestamp() != nil
	var conditions []v1.Condition
	if l.manageConditions {
		instanceConditionsObj, err := toRuntimeObjectConditionsInterface(instance, log)
		if err != nil {
			return ctrl.Result{}, err
		}
		conditions = instanceConditionsObj.GetConditions()
	}

	if l.spreadReconciles && instance.GetDeletionTimestamp().IsZero() {
		instanceStatusObj := MustToRuntimeObjectSpreadReconcileStatusInterface(instance, log)
		generationIsDifferent := instance.GetGeneration() != instanceStatusObj.GetObservedGeneration()
		isAfterNextReconcileTime := v1.Now().UTC().After(instanceStatusObj.GetNextReconcileTime().Time.UTC())
		refreshRequested := slices.Contains(maps.Keys(instance.GetLabels()), SpreadReconcileRefreshLabel)

		reconcileRequired := generationIsDifferent || isAfterNextReconcileTime || refreshRequested
		if !reconcileRequired {
			return onNextReconcile(instanceStatusObj, log)
		}
	}

	if l.manageConditions {
		setInstanceConditionUnknownIfNotSet(&conditions)
	}

	if l.prepareContextFunc != nil {
		localCtx, oErr := l.prepareContextFunc(ctx, instance)
		if oErr != nil {
			return l.handleOperatorError(ctx, oErr, "failed to prepare context")
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
			instanceConditionsObj, err := toRuntimeObjectConditionsInterface(instance, log)
			if err != nil {
				return ctrl.Result{}, err
			}
			instanceConditionsObj.SetConditions(conditions)
		}
		subResult, retry, err := l.reconcileSubroutine(ctx, instance, subroutine, log, sentryTags)
		// Update conditions with any changes the subroutine did
		if l.manageConditions {
			instanceConditionsObj, err := toRuntimeObjectConditionsInterface(instance, log)
			if err != nil {
				return ctrl.Result{}, err
			}
			conditions = instanceConditionsObj.GetConditions()
		}
		if err != nil {
			if l.manageConditions {
				setSubroutineCondition(&conditions, subroutine, result, err, inDeletion, log)
				setInstanceConditionReady(&conditions, v1.ConditionFalse)
				instanceConditionsObj, err := toRuntimeObjectConditionsInterface(instance, log)
				if err != nil {
					return ctrl.Result{}, err
				}
				instanceConditionsObj.SetConditions(conditions)
			}
			if !retry {
				err := l.markResourceAsFinal(instance, log, conditions, v1.ConditionFalse)
				if err != nil {
					return ctrl.Result{}, err
				}
			}
			_ = updateStatus(ctx, l.client, originalCopy, instance, log, sentryTags)
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
		err := l.markResourceAsFinal(instance, log, conditions, v1.ConditionTrue)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if l.manageConditions {
			setInstanceConditionReady(&conditions, v1.ConditionFalse)
		}
	}

	if l.manageConditions {
		instanceConditionsObj, err := toRuntimeObjectConditionsInterface(instance, log)
		if err != nil {
			return ctrl.Result{}, err
		}
		instanceConditionsObj.SetConditions(conditions)
	}

	err = updateStatus(ctx, l.client, originalCopy, instance, log, sentryTags)
	if err != nil {
		return result, err
	}

	if l.spreadReconciles && instance.GetDeletionTimestamp().IsZero() {
		removed := removeRefreshLabelIfExists(instance)
		if removed {
			updateErr := l.client.Update(ctx, instance)
			if updateErr != nil {
				return l.handleClientError("failed to update instance", log, err, sentryTags)
			}
		}
	}

	log.Info().Msg("end reconcile")
	return result, nil
}

func (l *LifecycleManager) markResourceAsFinal(instance RuntimeObject, log *logger.Logger, conditions []v1.Condition, status v1.ConditionStatus) error {
	if l.spreadReconciles && instance.GetDeletionTimestamp().IsZero() {
		instanceStatusObj := MustToRuntimeObjectSpreadReconcileStatusInterface(instance, log)
		setNextReconcileTime(instanceStatusObj, log)
		updateObservedGeneration(instanceStatusObj, log)
	}

	if l.manageConditions {
		setInstanceConditionReady(&conditions, status)
	}
	return nil
}

func updateStatus(ctx context.Context, cl client.Client, original runtime.Object, current RuntimeObject, log *logger.Logger, sentryTags sentry.Tags) error {

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
			sentry.CaptureError(err, sentryTags, sentry.Extras{"message": "Updating of instance status failed"})
		}
		log.Error().Err(err).Msg("cannot update reconciliation Conditions, kubernetes client error")
		return err
	}

	return nil
}

func (l *LifecycleManager) handleOperatorError(ctx context.Context, operatorError errors.OperatorError, msg string) (ctrl.Result, error) {
	l.log.Error().Bool("retry", operatorError.Retry()).Bool("sentry", operatorError.Sentry()).Err(operatorError.Err()).Msg(msg)
	if operatorError.Sentry() {
		sentry.CaptureError(operatorError.Err(), sentry.GetSentryTagsFromContext(ctx))
	}

	if operatorError.Retry() {
		return ctrl.Result{}, operatorError.Err()
	}

	return ctrl.Result{}, nil
}

func (l *LifecycleManager) handleClientError(msg string, log *logger.Logger, err error, sentryTags sentry.Tags) (ctrl.Result, error) {
	log.Error().Err(err).Msg(msg)
	sentry.CaptureError(err, sentryTags)
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

func (l *LifecycleManager) reconcileSubroutine(ctx context.Context, instance RuntimeObject, subroutine Subroutine, log *logger.Logger, sentryTags map[string]string) (ctrl.Result, bool, error) {
	subroutineLogger := log.ChildLogger("subroutine", subroutine.GetName())
	ctx = logger.SetLoggerInContext(ctx, subroutineLogger)
	subroutineLogger.Debug().Msg("start subroutine")

	ctx, span := otel.Tracer(l.operatorName).Start(ctx, fmt.Sprintf("%s.reconcileSubroutine.%s", l.controllerName, subroutine.GetName()))
	defer span.End()
	var result ctrl.Result
	var err errors.OperatorError
	if instance.GetDeletionTimestamp() != nil {
		if containsFinalizer(instance, subroutine.Finalizers()) {
			result, err = subroutine.Finalize(ctx, instance)
			if err == nil {
				// Remove finalizers unless requeue is requested
				err = l.removeFinalizerIfNeeded(ctx, instance, subroutine, result)
			}
		}
	} else {
		err = l.addFinalizerIfNeeded(ctx, instance, subroutine)
		if err == nil {
			result, err = subroutine.Process(ctx, instance)
		}
	}
	if err != nil {
		if err.Sentry() {
			log.Error().Err(err.Err()).Msg("subroutine ended with error")
			sentry.CaptureError(err.Err(), sentryTags)
		}
		subroutineLogger.Error().Err(err.Err()).Bool("retry", err.Retry()).Msg("subroutine ended with error")
		return result, err.Retry(), err.Err()
	}

	subroutineLogger.Debug().Msg("end subroutine")
	return result, false, nil
}

func (l *LifecycleManager) removeFinalizerIfNeeded(ctx context.Context, instance RuntimeObject, subroutine Subroutine, result ctrl.Result) errors.OperatorError {
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

func (l *LifecycleManager) addFinalizerIfNeeded(ctx context.Context, instance RuntimeObject, subroutine Subroutine) errors.OperatorError {
	update := false
	for _, f := range subroutine.Finalizers() {
		needsUpdate := controllerutil.AddFinalizer(instance, f)
		if needsUpdate {
			update = true
		}
	}
	if update {
		// When calling Update on the resource, any unsaved changes to the status like conditions will be lost. Let's keep the conditions before updating
		var conditions []v1.Condition
		if l.manageConditions {
			instanceConditionsObj, err := toRuntimeObjectConditionsInterface(instance, l.log)
			if err != nil {
				return errors.NewOperatorError(errors.Wrap(err, "failed to get conditions"), true, false)
			}
			conditions = instanceConditionsObj.GetConditions()
		}

		updateErr := l.client.Update(ctx, instance)
		if updateErr != nil {
			return errors.NewOperatorError(errors.Wrap(updateErr, "failed to update instance"), true, false)
		}
		if l.manageConditions {
			instanceConditionsObj, err := toRuntimeObjectConditionsInterface(instance, l.log)
			if err != nil {
				return errors.NewOperatorError(errors.Wrap(err, "failed to get conditions"), true, false)
			}
			instanceConditionsObj.SetConditions(conditions)
		}
	}
	return nil
}

func (l *LifecycleManager) SetupWithManagerBuilder(mgr ctrl.Manager, maxReconciles int, reconcilerName string, instance RuntimeObject, debugLabelValue string, log *logger.Logger, eventPredicates ...predicate.Predicate) (*builder.Builder, error) {
	if l.manageConditions {
		_, err := toRuntimeObjectConditionsInterface(instance, log)
		if err != nil {
			return nil, err
		}
	}

	if l.spreadReconciles {
		_, err := toRuntimeObjectSpreadReconcileStatusInterface(instance, log)
		if err != nil {
			return nil, err
		}
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
