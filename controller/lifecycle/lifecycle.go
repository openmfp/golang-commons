package lifecycle

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"golang.org/x/exp/maps"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
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
	log              *logger.Logger
	client           client.Client
	subroutines      []Subroutine
	operatorName     string
	controllerName   string
	spreadReconciles bool
	manageConditions bool
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

	log, err := l.log.ChildLoggerWithAttributes("name", req.Name, "namespace", req.Namespace, "reconcile_id", reconcileId)
	if err != nil {
		return ctrl.Result{}, err
	}

	sentryTags := sentry.Tags{"namespace": req.Namespace, "name": req.Name}

	ctx = logger.SetLoggerInContext(ctx, log)
	ctx = sentry.ContextWithSentryTags(ctx, sentryTags)

	log.Info().Msg("start reconcile")
	err = l.client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			log.Info().Msg("instance not found. It was likely deleted")
			return ctrl.Result{}, nil
		}
		return l.handleClientError("failed to retrieve instance", log, err, sentryTags)
	}

	originalCopy := instance.DeepCopyObject()

	if l.spreadReconciles && instance.GetDeletionTimestamp().IsZero() {
		if instanceStatusObj, ok := instance.(RuntimeObjectSpreadReconcileStatus); ok {
			if !slices.Contains(maps.Keys(instance.GetLabels()), SpreadReconcileRefreshLabel) &&
				(instance.GetGeneration() == instanceStatusObj.GetObservedGeneration() || v1.Now().UTC().Before(instanceStatusObj.GetNextReconcileTime().Time.UTC())) {
				return onNextReconcile(instanceStatusObj, log)
			}
		} else {
			err = fmt.Errorf("spreadReconciles is enabled, but instance does not implement RuntimeObjectSpreadReconcileStatus interface. This is a programming error")
			log.Error().Err(err).Msg("Error during reconcile")
			sentry.CaptureError(err, sentryTags)
			return ctrl.Result{}, err
		}
	}

	if l.manageConditions {
		if instanceConditionsObj, ok := instance.(RuntimeObjectConditions); ok {
			conditions := instanceConditionsObj.GetConditions()
			if setReady(&conditions, v1.ConditionUnknown) {
				instanceConditionsObj.SetConditions(conditions)
			}
		} else {
			err = fmt.Errorf("manageConditions is enabled, but instance does not implement RuntimeObjectConditions interface. This is a programming error")
			log.Error().Err(err).Msg("Error during reconcile")
			sentry.CaptureError(err, sentryTags)
			return ctrl.Result{}, err
		}
	}
	// Continue with reconciliation
	for _, subroutine := range l.subroutines {
		if l.manageConditions {
			if instanceConditionsObj, ok := instance.(RuntimeObjectConditions); ok {
				conditions := instanceConditionsObj.GetConditions()
				if setSubroutineCondition(&conditions, subroutine.GetName(), v1.ConditionUnknown, "The subroutine is being processed", "SubroutineProcessing") {
					instanceConditionsObj.SetConditions(conditions)
				}
			}
		}
		subResult, err := l.reconcileSubroutine(ctx, instance, subroutine, log, sentryTags)
		if err != nil {
			if l.manageConditions {
				if instanceConditionsObj, ok := instance.(RuntimeObjectConditions); ok {
					conditions := instanceConditionsObj.GetConditions()
					if setSubroutineCondition(&conditions, subroutine.GetName(), v1.ConditionFalse, "The subroutine failed", "SubroutineFailed") {
						instanceConditionsObj.SetConditions(conditions)
					}
				}
			}
			_ = l.updateStatus(ctx, originalCopy, instance, log, sentryTags)
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
			if instanceConditionsObj, ok := instance.(RuntimeObjectConditions); ok {
				conditions := instanceConditionsObj.GetConditions()
				if !subResult.Requeue && subResult.RequeueAfter == 0 {
					// Subroutine was successful
					if setSubroutineCondition(&conditions, subroutine.GetName(), v1.ConditionTrue, "The subroutine was successful", "SubroutineSuccess") {
						instanceConditionsObj.SetConditions(conditions)
					}
				}
			}
		}
	}

	if !result.Requeue && result.RequeueAfter == 0 {
		// Reconciliation was successful
		if l.spreadReconciles && instance.GetDeletionTimestamp().IsZero() {
			if instanceStatusObj, ok := instance.(RuntimeObjectSpreadReconcileStatus); ok {
				setNextReconcileTime(instanceStatusObj, log)
				updateObservedGeneration(instanceStatusObj, log)
			}
		}

		if l.manageConditions {
			if instanceConditionsObj, ok := instance.(RuntimeObjectConditions); ok {
				conditions := instanceConditionsObj.GetConditions()
				if setReady(&conditions, v1.ConditionTrue) {
					instanceConditionsObj.SetConditions(conditions)
				}
			}
		}
	} else {
		if l.manageConditions {
			if instanceConditionsObj, ok := instance.(RuntimeObjectConditions); ok {
				conditions := instanceConditionsObj.GetConditions()
				if setReady(&conditions, v1.ConditionFalse) {
					instanceConditionsObj.SetConditions(conditions)
				}
			}
		}
	}

	err = l.updateStatus(ctx, originalCopy, instance, log, sentryTags)
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

func (l *LifecycleManager) updateStatus(ctx context.Context, original runtime.Object, current RuntimeObject, log *logger.Logger, sentryTags sentry.Tags) error {
	currentStatus := reflect.Indirect(reflect.ValueOf(current)).FieldByName("Status").Interface()
	originalStatus := reflect.Indirect(reflect.ValueOf(original)).FieldByName("Status").Interface()
	equal := reflect.DeepEqual(currentStatus, originalStatus)
	if !equal {
		log.Info().Msg("updating resource status")
		err := l.client.Status().Update(ctx, current)
		if err != nil {
			if !k8sErrors.IsConflict(err) {
				sentry.CaptureError(err, sentryTags, sentry.Extras{"message": "Updating of instance status failed"})
			}
			log.Error().Err(err).Msg("cannot update reconciliation Conditions, kubernetes client error")
			return err
		}
	} else {
		log.Info().Msg("skipping status update, since they are equal")
	}
	return nil
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

func (l *LifecycleManager) reconcileSubroutine(ctx context.Context, instance RuntimeObject, subroutine Subroutine, log *logger.Logger, sentryTags map[string]string) (ctrl.Result, error) {
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
			// Remove finalizers unless requeue is requested
			err = l.removeFinalizerIfNeeded(ctx, instance, subroutine, err, result)
		}
	} else {
		err = l.addFinalizerIfNeeded(ctx, instance, subroutine)
		if err == nil {
			result, err = subroutine.Process(ctx, instance)
		}
	}
	if err != nil && err.Sentry() {
		sentry.CaptureError(err.Err(), sentryTags)
	}
	if err != nil && err.Retry() {
		subroutineLogger.Error().Err(err.Err()).Msg("subroutine ended with error")
		return result, err.Err()
	}
	subroutineLogger.Debug().Msg("end subroutine")
	return result, nil
}

func (l *LifecycleManager) removeFinalizerIfNeeded(ctx context.Context, instance RuntimeObject, subroutine Subroutine, err errors.OperatorError, result ctrl.Result) errors.OperatorError {
	if err == nil && !result.Requeue && result.RequeueAfter == 0 {
		update := false
		for _, f := range subroutine.Finalizers() {
			needsUpdate := controllerutil.RemoveFinalizer(instance, f)
			if needsUpdate {
				update = true
			}
		}
		if update {
			updateErr := l.client.Update(ctx, instance)
			if updateErr != nil {
				return errors.NewOperatorError(errors.Wrap(updateErr, "failed to update instance"), true, false)
			}
		}
	}
	return err
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
		updateErr := l.client.Update(ctx, instance)
		if updateErr != nil {
			return errors.NewOperatorError(errors.Wrap(updateErr, "failed to update instance"), true, false)
		}
	}
	return nil
}

func (l *LifecycleManager) SetupWithManager(mgr ctrl.Manager, maxReconciles int, reconcilerName string, instance RuntimeObject, debugLabelValue string, r reconcile.Reconciler, eventPredicates ...predicate.Predicate) error {
	eventPredicates = append([]predicate.Predicate{filter.DebugResourcesBehaviourPredicate(debugLabelValue)}, eventPredicates...)
	return ctrl.NewControllerManagedBy(mgr).
		Named(reconcilerName).
		For(instance).
		WithOptions(controller.Options{MaxConcurrentReconciles: maxReconciles}).
		WithEventFilter(predicate.And(eventPredicates...)).
		Complete(r)
}
