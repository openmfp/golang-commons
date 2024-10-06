package lifecycle

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/openmfp/golang-commons/logger"
	"github.com/openmfp/golang-commons/sentry"
)

const (
	ConditionReady = "Ready"

	messageResourceReady      = "The resource is ready"
	messageResourceNotReady   = "The resource is not ready"
	messageResourceProcessing = "The resource is processing"

	reasonComplete   = "Complete"
	reasonProcessing = "Processing"
	reasonError      = "Error"

	subroutineReadyConditionFormatString    = "%s_Ready"
	subroutineFinalizeConditionFormatString = "%s_Finalize"

	subroutineMessageProcessingFormatString = "The %s is processing"
	subroutineMessageCompleteFormatString   = "The %s is complete"
	subroutineMessageErrorFormatString      = "The %s has an error: %s"
)

func (l *LifecycleManager) WithConditionManagement() *LifecycleManager {
	l.manageConditions = true
	return l
}

type RuntimeObjectConditions interface {
	GetConditions() []metav1.Condition
	SetConditions([]metav1.Condition)
}

// Set the Condition of the instance to be ready
func setInstanceConditionReady(conditions *[]metav1.Condition, status metav1.ConditionStatus) bool {
	var msg string
	switch status {
	case metav1.ConditionTrue:
		msg = messageResourceReady
	case metav1.ConditionFalse:
		msg = messageResourceNotReady
	default:
		msg = messageResourceProcessing
	}
	return meta.SetStatusCondition(conditions, metav1.Condition{
		Type:    ConditionReady,
		Status:  status,
		Message: msg,
		Reason:  reasonComplete,
	})
}

// Set the Condition to be Unknown in case it is not set yet
func setInstanceConditionUnknownIfNotSet(conditions *[]metav1.Condition) bool {
	existingCondition := meta.FindStatusCondition(*conditions, ConditionReady)
	if existingCondition == nil {
		return setInstanceConditionReady(conditions, metav1.ConditionUnknown)
	}
	return false
}

func setSubroutineConditionToUnknownIfNotSet(conditions *[]metav1.Condition, subroutine Subroutine, isFinalize bool, log *logger.Logger) bool {
	conditionName, conditionMessage := getConditionNameAndMessage(subroutine, isFinalize)

	existingCondition := meta.FindStatusCondition(*conditions, conditionName)
	if existingCondition == nil {
		changed := meta.SetStatusCondition(conditions,
			metav1.Condition{Type: conditionName, Status: metav1.ConditionUnknown, Message: fmt.Sprintf(subroutineMessageProcessingFormatString, conditionMessage), Reason: reasonProcessing})
		if changed {
			log.Info().Str("type", conditionName).Msg("updated condition")
		}
		return changed
	}
	return false
}

func getConditionNameAndMessage(subroutine Subroutine, isFinalize bool) (string, string) {
	conditionName := fmt.Sprintf(subroutineReadyConditionFormatString, subroutine.GetName())
	conditionMessage := "subroutine"
	if isFinalize {
		conditionName = fmt.Sprintf(subroutineFinalizeConditionFormatString, subroutine.GetName())
		conditionMessage = "subroutine finalization"
	}
	return conditionName, conditionMessage
}

// Set Subroutines Conditions
func setSubroutineCondition(conditions *[]metav1.Condition, subroutine Subroutine, subroutineResult ctrl.Result, subroutineErr error, isFinalize bool, log *logger.Logger) bool {
	conditionName, conditionMessage := getConditionNameAndMessage(subroutine, isFinalize)

	// processing complete
	if subroutineErr == nil && !subroutineResult.Requeue && subroutineResult.RequeueAfter == 0 {
		return meta.SetStatusCondition(conditions,
			metav1.Condition{Type: conditionName, Status: metav1.ConditionTrue, Message: fmt.Sprintf(subroutineMessageCompleteFormatString, conditionMessage), Reason: reasonComplete})
	}
	// processing is still processing
	if subroutineErr == nil && (subroutineResult.RequeueAfter > 0 || subroutineResult.Requeue) {
		return meta.SetStatusCondition(conditions,
			metav1.Condition{Type: conditionName, Status: metav1.ConditionUnknown, Message: fmt.Sprintf(subroutineMessageProcessingFormatString, conditionMessage), Reason: reasonProcessing})
	}
	// processing failed
	var sErr error
	if subroutineErr != nil {
		sErr = subroutineErr
	}
	changed := meta.SetStatusCondition(conditions,
		metav1.Condition{Type: conditionName, Status: metav1.ConditionFalse, Message: fmt.Sprintf(subroutineMessageErrorFormatString, conditionMessage, sErr), Reason: reasonError})
	if changed {
		log.Info().Str("type", conditionName).Msg("updated condition")
	}
	return changed
}

func toRuntimeObjectConditionsInterface(instance RuntimeObject, log *logger.Logger) (RuntimeObjectConditions, error) {
	if obj, ok := instance.(RuntimeObjectConditions); ok {
		return obj, nil
	}
	err := fmt.Errorf("manageConditions is enabled, but instance does not implement RuntimeObjectConditions interface. This is a programming error")
	log.Error().Err(err).Msg("instance does not implement RuntimeObjectConditions interface")
	sentry.CaptureError(err, nil)
	return nil, err
}

func MustToRuntimeObjectConditionsInterface(instance RuntimeObject, log *logger.Logger) RuntimeObjectConditions {
	obj, err := toRuntimeObjectConditionsInterface(instance, log)
	if err == nil {
		return obj
	}
	log.Panic().Err(err).Msg("instance does not implement RuntimeObjectConditions interface")
	return nil
}
