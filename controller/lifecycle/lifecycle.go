package lifecycle

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
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
	logger           *logger.Logger
	client           client.Client
	subroutines      []Subroutine
	operatorName     string
	controllerName   string
	spreadReconciles bool
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

func NewLifecycleManager(logger *logger.Logger, operatorName string, controllerName string, client client.Client, subroutines []Subroutine) *LifecycleManager {
	return &LifecycleManager{
		logger:           logger,
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
	log := logger.NewFromZerolog(l.logger.With().Str("name", req.Name).Str("namespace", req.Namespace).Str("reconcile_id", reconcileId).Logger())
	sentryTags := sentry.Tags{"namespace": req.Namespace, "name": req.Name}

	ctx = logger.SetLoggerInContext(ctx, log)
	ctx = sentry.ContextWithSentryTags(ctx, sentryTags)

	log.Info().Msg("start reconcile")
	err := l.client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			log.Info().Msg("instance not found. It was likely deleted")
			return ctrl.Result{}, nil
		}

		log.Error().Err(err).Msg("failed to retrieve instance")
		sentry.CaptureError(err, sentryTags)
		return ctrl.Result{}, err
	}

	c := instance.DeepCopyObject()

	if l.spreadReconciles {
		if instanceStatusObj, ok := instance.(RuntimeObjectSpreadReconcileStatus); ok {
			if instance.GetGeneration() == instanceStatusObj.GetObservedGeneration() || v1.Now().UTC().Before(instanceStatusObj.GetNextReconcileTime().Time.UTC()) {
				return onNextReconcile(instanceStatusObj, log)
			}
		} else {
			err = fmt.Errorf("spreadReconciles is enabled, but instance does not implement RuntimeObjectSpreadReconcileStatus interface. This is a programming error")
			log.Error().Err(err).Msg("Error during reconcile")
			sentry.CaptureError(err, sentryTags)
			return ctrl.Result{}, err
		}
	}
	// Continue with reconciliation
	for _, subroutine := range l.subroutines {
		subResult, err := l.reconcileSubroutine(ctx, instance, subroutine, log, sentryTags)
		if err != nil {
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
	}

	if !result.Requeue && result.RequeueAfter == 0 {
		// Reconciliation was successful
		if l.spreadReconciles {
			if instanceStatusObj, ok := instance.(RuntimeObjectSpreadReconcileStatus); ok {
				setNextReconcileTime(instanceStatusObj, log)
				updateObservedGeneration(instanceStatusObj, log)
			}
		}
	}

	currentStatus := reflect.Indirect(reflect.ValueOf(instance)).FieldByName("Status").Interface()
	originalStatus := reflect.Indirect(reflect.ValueOf(c)).FieldByName("Status").Interface()
	equal := reflect.DeepEqual(currentStatus, originalStatus)
	if !equal {
		log.Info().Msg("updating resource status")
		err = l.client.Status().Update(ctx, instance)
		if err != nil {
			if !k8sErrors.IsConflict(err) {
				sentry.CaptureError(err, sentryTags, sentry.Extras{"message": "Updating of instance status failed"})
			}
			log.Error().Err(err).Msg("cannot update reconciliation Conditions, kubernetes client error")
			return result, err
		}
	} else {
		log.Info().Msg("skipping status update, since they are equal")
	}

	log.Info().Msg("end reconcile")
	return result, nil
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
	subroutineLogger := logger.NewFromZerolog(log.With().Str("subroutine", subroutine.GetName()).Logger())
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
