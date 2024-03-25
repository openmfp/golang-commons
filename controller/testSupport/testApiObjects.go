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
func (m *TestApiObject) GetGeneration() int64 {
	return m.Generation
}

func (m *TestApiObject) GetObservedGeneration() int64 {
	return m.Status.ObservedGeneration
}

func (m *TestApiObject) SetObservedGeneration(g int64) {
	m.Status.ObservedGeneration = g
}

func (m *TestApiObject) GetNextReconcileTime() metav1.Time {
	return m.Status.NextReconcileTime
}

func (m *TestApiObject) SetNextReconcileTime(time metav1.Time) {
	m.Status.NextReconcileTime = time
}
