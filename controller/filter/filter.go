package filter

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	DebugLabel = "debug.openmfp.io"
)

// DebugResourcesBehaviourPredicate returns whether a resource should be digested
// depending on whether the DebugLabel matches the compareValue.
// To match resources where the label is not set, provide an empty string.
// This should be the default production configuration which can be overwritten for local development.
func DebugResourcesBehaviourPredicate(labelValue string) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			val := e.Object.GetLabels()[DebugLabel]
			return val == labelValue
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			val := e.ObjectNew.GetLabels()[DebugLabel]
			return val == labelValue
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			val := e.Object.GetLabels()[DebugLabel]
			return val == labelValue
		},
		GenericFunc: func(e event.GenericEvent) bool {
			val := e.Object.GetLabels()[DebugLabel]
			return val == labelValue
		},
	}
}
