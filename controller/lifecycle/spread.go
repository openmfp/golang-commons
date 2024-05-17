package lifecycle

import (
	"fmt"
	"math/rand"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/openmfp/golang-commons/logger"
	"github.com/openmfp/golang-commons/sentry"
)

const SpreadReconcileRefreshLabel = "openmfp.io/refresh-reconcile"

// WithSpreadingReconciles sets the LifecycleManager to spread out the reconciles
func (l *LifecycleManager) WithSpreadingReconciles() *LifecycleManager {
	l.spreadReconciles = true
	return l
}

type RuntimeObjectSpreadReconcileStatus interface {
	GetGeneration() int64
	GetObservedGeneration() int64
	SetObservedGeneration(int64)
	GetNextReconcileTime() v1.Time
	SetNextReconcileTime(time v1.Time)
}

// getNextReconcileTime returns a random time between 12 and 24 hours
func getNextReconcileTime() time.Duration {
	return 12*time.Hour + time.Duration(rand.Int63n(12*60))*time.Minute
}

// onNextReconcile is a helper function to set the next reconcile time and return the requeueAfter time
func onNextReconcile(instanceStatusObj RuntimeObjectSpreadReconcileStatus, log *logger.Logger) (ctrl.Result, error) {
	requeueAfter := time.Until(instanceStatusObj.GetNextReconcileTime().Time.UTC())
	log.Debug().Int64("minutes-till-next-execution", int64(requeueAfter.Minutes())).Msg("Completed reconciliation, no processing needed")
	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

// setNextReconcileTime calculates and sets the next reconcile time for the instance
func setNextReconcileTime(instanceStatusObj RuntimeObjectSpreadReconcileStatus, log *logger.Logger) {
	nextReconcileTime := getNextReconcileTime()
	log.Debug().Int64("minutes-till-next-execution", int64(nextReconcileTime.Minutes())).Msg("Setting next reconcile time for the instance")
	instanceStatusObj.SetNextReconcileTime(v1.NewTime(time.Now().Add(nextReconcileTime)))
}

// updateObservedGeneration updates the observed generation of the instance struct
func updateObservedGeneration(instanceStatusObj RuntimeObjectSpreadReconcileStatus, log *logger.Logger) {
	log.Debug().Int64("observed-generation", instanceStatusObj.GetObservedGeneration()).Int64("generation", instanceStatusObj.GetGeneration()).Msg("Updating observed generation")
	instanceStatusObj.SetObservedGeneration(instanceStatusObj.GetGeneration())
}
func removeRefreshLabelIfExists(instance RuntimeObject) bool {
	keyCount := len(instance.GetLabels())
	delete(instance.GetLabels(), SpreadReconcileRefreshLabel)
	return keyCount != len(instance.GetLabels())
}

func toRuntimeObjectSpreadReconcileStatusInterface(instance RuntimeObject, log *logger.Logger) (RuntimeObjectSpreadReconcileStatus, error) {
	if obj, ok := instance.(RuntimeObjectSpreadReconcileStatus); ok {
		return obj, nil
	}
	err := fmt.Errorf("spreadReconciles is enabled, but instance does not implement RuntimeObjectSpreadReconcileStatus interface. This is a programming error")
	log.Error().Err(err).Msg("Failed to cast instance to RuntimeObjectSpreadReconcileStatus")
	sentry.CaptureError(err, nil)
	return nil, err
}

func MustToRuntimeObjectSpreadReconcileStatusInterface(instance RuntimeObject, log *logger.Logger) RuntimeObjectSpreadReconcileStatus {
	if obj, ok := instance.(RuntimeObjectSpreadReconcileStatus); ok {
		return obj
	}
	err := fmt.Errorf("spreadReconciles is enabled, but instance does not implement RuntimeObjectSpreadReconcileStatus interface. This is a programming error")
	sentry.CaptureError(err, nil)
	log.Panic().Msg("Failed to cast instance to RuntimeObjectSpreadReconcileStatus")
	return nil
}
