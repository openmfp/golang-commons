package lifecycle

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ConditionReady = "Ready"
)

func (l *LifecycleManager) WithConditionManagement() *LifecycleManager {
	l.manageConditions = true
	return l
}

type RuntimeObjectConditions interface {
	GetConditions() []metav1.Condition
	SetConditions([]metav1.Condition)
}

func setReady(conditions *[]metav1.Condition, status metav1.ConditionStatus) bool {

	var msg string
	switch status {
	case metav1.ConditionTrue:
		msg = "The resource is ready"
	case metav1.ConditionFalse:
		msg = "The resource is not ready"
	default:
		msg = ""
	}
	return meta.SetStatusCondition(conditions, metav1.Condition{
		Type:    ConditionReady,
		Status:  status,
		Message: msg,
		Reason:  ConditionReady,
	})
}

// Function to set Ready Condition to unknown in case it is not set or not ready
func setUnknownIfNotSet(conditions *[]metav1.Condition) bool {
	existingCondition := meta.FindStatusCondition(*conditions, ConditionReady)
	if existingCondition == nil {
		return setReady(conditions, metav1.ConditionUnknown)
	}
	return false
}

func setSubroutineCondition(conditions *[]metav1.Condition, subroutineName string, status metav1.ConditionStatus, message string, reason string) bool {
	name := fmt.Sprintf("%s_Ready", subroutineName)
	return meta.SetStatusCondition(conditions, metav1.Condition{
		Type:    name,
		Status:  status,
		Message: message,
		Reason:  reason,
	})
}
