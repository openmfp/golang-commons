package testSupport

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type TestApiObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status TestStatus `json:"status,omitempty"`
}
type TestStatus struct {
	Some               string
	NextReconcileTime  metav1.Time
	ObservedGeneration int64
}

func (t *TestApiObject) DeepCopyObject() runtime.Object {
	if c := t.DeepCopy(); c != nil {
		return c
	}
	return nil
}
func (t *TestApiObject) DeepCopy() *TestApiObject {
	if t == nil {
		return nil
	}
	out := new(TestApiObject)
	t.DeepCopyInto(out)
	return out
}
func (m *TestApiObject) DeepCopyInto(out *TestApiObject) {
	*out = *m
	out.TypeMeta = m.TypeMeta
	m.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
}
