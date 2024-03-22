package conditions

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ConditionsService interface {
	ConditionsSetter
	ConditionsGetter
}

type ConditionsSetter interface {
	SetTrue(objectMeta metav1.ObjectMeta, conditions *[]metav1.Condition, reason, message string)
	SetFalse(objectMeta metav1.ObjectMeta, conditions *[]metav1.Condition, reason, message string)
}

type ConditionsGetter interface {
	GetStatus(conditions []metav1.Condition) *metav1.Condition
	IsStatusTrue(conditions []metav1.Condition) bool
}
