package lifecycle

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmfp/golang-commons/errors"
)

type testApiObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status TestStatus `json:"status,omitempty"`
}

func (t *testApiObject) DeepCopyObject() runtime.Object {
	if c := t.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (t *testApiObject) DeepCopy() *testApiObject {
	if t == nil {
		return nil
	}
	out := new(testApiObject)
	t.DeepCopyInto(out)
	return out
}
func (m *testApiObject) DeepCopyInto(out *testApiObject) {
	*out = *m
	out.TypeMeta = m.TypeMeta
	m.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Status = m.Status
}

type notImplementingSpreadReconciles struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status TestStatus `json:"status,omitempty"`
}

func (m *notImplementingSpreadReconciles) DeepCopyObject() runtime.Object {
	if c := m.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (m *notImplementingSpreadReconciles) DeepCopy() *notImplementingSpreadReconciles {
	if m == nil {
		return nil
	}
	out := new(notImplementingSpreadReconciles)
	m.DeepCopyInto(out)
	return out
}
func (m *notImplementingSpreadReconciles) DeepCopyInto(out *notImplementingSpreadReconciles) {
	*out = *m
	out.TypeMeta = m.TypeMeta
	m.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Status = m.Status
}

type changeStatusSubroutine struct {
	client client.Client
}

func (c changeStatusSubroutine) Process(_ context.Context, runtimeObj RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	instance := runtimeObj.(*testApiObject)
	instance.Status.Some = "other string"
	return controllerruntime.Result{}, nil
}

func (c changeStatusSubroutine) Finalize(_ context.Context, _ RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	//TODO implement me
	panic("implement me")
}

func (c changeStatusSubroutine) GetName() string {
	return "changeStatus"
}

func (c changeStatusSubroutine) Finalizers() []string {
	return []string{}
}

type failureScenarioSubroutine struct {
	Retry      bool
	RequeAfter bool
}

func (f failureScenarioSubroutine) Process(_ context.Context, _ RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	if f.Retry {
		return controllerruntime.Result{Requeue: true}, nil
	}

	if f.RequeAfter {
		return controllerruntime.Result{RequeueAfter: 10 * time.Second}, nil
	}

	return controllerruntime.Result{}, errors.NewOperatorError(fmt.Errorf("failureScenarioSubroutine"), true, false)
}

func (f failureScenarioSubroutine) Finalize(_ context.Context, _ RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	return controllerruntime.Result{}, nil
}

func (f failureScenarioSubroutine) Finalizers() []string {
	return []string{}
}

func (c failureScenarioSubroutine) GetName() string {
	return "failureScenarioSubroutine"
}

type TestStatus struct {
	Some               string
	NextReconcileTime  metav1.Time
	ObservedGeneration int64
}

func (m *testApiObject) GetGeneration() int64 {
	return m.Generation
}

func (m *testApiObject) GetObservedGeneration() int64 {
	return m.Status.ObservedGeneration
}

func (m *testApiObject) SetObservedGeneration(g int64) {
	m.Status.ObservedGeneration = g
}

func (m *testApiObject) GetNextReconcileTime() v1.Time {
	return m.Status.NextReconcileTime
}

func (m *testApiObject) SetNextReconcileTime(time v1.Time) {
	m.Status.NextReconcileTime = time
}
