package lifecycle

import (
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
	return meta.SetStatusCondition(conditions, metav1.Condition{
		Type:    ConditionReady,
		Status:  status,
		Message: "The resource is ready",
		Reason:  ConditionReady,
	})
}
