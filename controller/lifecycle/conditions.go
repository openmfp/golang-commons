package lifecycle

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	ConditionReady    = "Ready"
	MessageComplete   = "Complete"
	MessageProcessing = "Processing"
	MessageError      = "Error"
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
		msg = "The resource is ready"
	case metav1.ConditionFalse:
		msg = "The resource is not ready"
	default:
		msg = "The resource is processing"
	}
	return meta.SetStatusCondition(conditions, metav1.Condition{
		Type:    ConditionReady,
		Status:  status,
		Message: msg,
		Reason:  ConditionReady,
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

func setSubroutineConditionToUnknownIfNotSet(conditions *[]metav1.Condition, subroutine Subroutine, isFinalize bool) bool {
	conditionName := fmt.Sprintf("%s_Ready", subroutine.GetName())
	if isFinalize {
		conditionName = fmt.Sprintf("%s_Finalize", subroutine.GetName())
	}
	existingCondition := meta.FindStatusCondition(*conditions, conditionName)
	if existingCondition == nil {
		changed := meta.SetStatusCondition(conditions,
			metav1.Condition{Type: conditionName, Status: metav1.ConditionUnknown, Message: "The subroutine finalization is processing", Reason: MessageProcessing})
		return changed

	}
	return false
}

// Set Subroutines Conditions
func setSubroutineCondition(conditions *[]metav1.Condition, subroutine Subroutine, subroutineResult ctrl.Result, subroutineErr error, isFinalize bool) bool {
	conditionName := fmt.Sprintf("%s_Ready", subroutine.GetName())
	conditionMessage := "subroutine"
	if isFinalize {
		conditionName = fmt.Sprintf("%s_Finalize", subroutine.GetName())
		conditionMessage = "subroutine finalization"
	}

	// processing complete
	if subroutineErr == nil && !subroutineResult.Requeue && subroutineResult.RequeueAfter == 0 {
		return meta.SetStatusCondition(conditions,
			metav1.Condition{Type: conditionName, Status: metav1.ConditionTrue, Message: fmt.Sprintf("The %s is complete", conditionMessage), Reason: MessageComplete})
	}
	// processing is still processing
	if subroutineErr == nil && (subroutineResult.RequeueAfter > 0 || subroutineResult.Requeue) {
		return meta.SetStatusCondition(conditions,
			metav1.Condition{Type: conditionName, Status: metav1.ConditionUnknown, Message: fmt.Sprintf("The %s is processing", conditionMessage), Reason: MessageProcessing})
	}
	// processing failed
	return meta.SetStatusCondition(conditions,
		metav1.Condition{Type: conditionName, Status: metav1.ConditionFalse, Message: fmt.Sprintf("The %s has an error: %s", conditionMessage, subroutineErr.Error()), Reason: MessageError})

}
