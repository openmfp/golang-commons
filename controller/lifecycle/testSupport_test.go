package lifecycle

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"

	"github.com/openmfp/golang-commons/context/keys"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmfp/golang-commons/controller/testSupport"
	"github.com/openmfp/golang-commons/errors"
)

const failureScenarioSubroutineFinalizer = "failuresubroutine"
const changeStatusSubroutineFinalizer = "changestatus"

type implementConditions struct {
	testSupport.TestApiObject `json:",inline"`
}

func (m *implementConditions) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}

func (m *implementConditions) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}

type implementingSpreadReconciles struct {
	testSupport.TestApiObject `json:",inline"`
}

func (m *implementingSpreadReconciles) GetGeneration() int64 {
	return m.Generation
}

func (m *implementingSpreadReconciles) GetObservedGeneration() int64 {
	return m.Status.ObservedGeneration
}

func (m *implementingSpreadReconciles) SetObservedGeneration(g int64) {
	m.Status.ObservedGeneration = g
}

func (m *implementingSpreadReconciles) GetNextReconcileTime() metav1.Time {
	return m.Status.NextReconcileTime
}

func (m *implementingSpreadReconciles) SetNextReconcileTime(time metav1.Time) {
	m.Status.NextReconcileTime = time
}

type notImplementingSpreadReconciles struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status testSupport.TestStatus `json:"status,omitempty"`
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
	if instance, ok := runtimeObj.(*testSupport.TestApiObject); ok {
		instance.Status.Some = "other string"
	}
	if instance, ok := runtimeObj.(*implementingSpreadReconciles); ok {
		instance.Status.Some = "other string"
	}

	if instance, ok := runtimeObj.(*implementConditions); ok {
		instance.Status.Some = "other string"
	}
	return controllerruntime.Result{}, nil
}

func (c changeStatusSubroutine) Finalize(_ context.Context, _ RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	return controllerruntime.Result{}, nil
}

func (c changeStatusSubroutine) GetName() string {
	return "changeStatus"
}

func (c changeStatusSubroutine) Finalizers() []string {
	return []string{"changestatus"}
}

type addConditionSubroutine struct {
	Ready metav1.ConditionStatus
}

func (c addConditionSubroutine) Process(_ context.Context, runtimeObj RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	if instance, ok := runtimeObj.(*implementConditions); ok {
		instance.Status.Some = "other string"
		meta.SetStatusCondition(&instance.Status.Conditions, metav1.Condition{
			Type:    "test",
			Status:  c.Ready,
			Reason:  "test",
			Message: "test",
		})
	}

	return controllerruntime.Result{}, nil
}

func (c addConditionSubroutine) Finalize(_ context.Context, _ RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	return controllerruntime.Result{}, nil
}

func (c addConditionSubroutine) GetName() string {
	return "addCondition"
}

func (c addConditionSubroutine) Finalizers() []string {
	return []string{}
}

type failureScenarioSubroutine struct {
	Retry              bool
	RequeAfter         bool
	FinalizeRetry      bool
	FinalizeRequeAfter bool
}

func (f failureScenarioSubroutine) Process(_ context.Context, _ RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	if f.RequeAfter {
		return controllerruntime.Result{RequeueAfter: 10 * time.Second}, nil
	}

	return controllerruntime.Result{}, errors.NewOperatorError(fmt.Errorf("failureScenarioSubroutine"), f.Retry, false)
}

func (f failureScenarioSubroutine) Finalize(_ context.Context, _ RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	if f.Retry {
		return controllerruntime.Result{Requeue: true}, nil
	}

	if f.RequeAfter {
		return controllerruntime.Result{RequeueAfter: 10 * time.Second}, nil
	}

	return controllerruntime.Result{}, errors.NewOperatorError(fmt.Errorf("failureScenarioSubroutine"), true, false)
}

func (f failureScenarioSubroutine) Finalizers() []string {
	return []string{failureScenarioSubroutineFinalizer}
}

func (c failureScenarioSubroutine) GetName() string {
	return "failureScenarioSubroutine"
}

type implementConditionsAndSpreadReconciles struct {
	testSupport.TestApiObject `json:",inline"`
}

func (m *implementConditionsAndSpreadReconciles) GetConditions() []metav1.Condition {
	return m.Status.Conditions
}
func (m *implementConditionsAndSpreadReconciles) SetConditions(conditions []metav1.Condition) {
	m.Status.Conditions = conditions
}
func (m *implementConditionsAndSpreadReconciles) GetGeneration() int64 {
	return m.Generation
}
func (m *implementConditionsAndSpreadReconciles) GetObservedGeneration() int64 {
	return m.Status.ObservedGeneration
}
func (m *implementConditionsAndSpreadReconciles) SetObservedGeneration(g int64) {
	m.Status.ObservedGeneration = g
}

func (m *implementConditionsAndSpreadReconciles) GetNextReconcileTime() metav1.Time {
	return m.Status.NextReconcileTime
}
func (m *implementConditionsAndSpreadReconciles) SetNextReconcileTime(time metav1.Time) {
	m.Status.NextReconcileTime = time
}

type contextValueSubroutine struct {
}

const contextValueKey = keys.ContextKey("contextValueKey")

func (f contextValueSubroutine) Process(ctx context.Context, r RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	if instance, ok := r.(*testSupport.TestApiObject); ok {
		instance.Status.Some = ctx.Value(contextValueKey).(string)
	}
	return controllerruntime.Result{}, nil
}

func (f contextValueSubroutine) Finalize(_ context.Context, _ RuntimeObject) (controllerruntime.Result, errors.OperatorError) {
	return controllerruntime.Result{}, nil
}

func (f contextValueSubroutine) Finalizers() []string {
	return []string{}
}

func (c contextValueSubroutine) GetName() string {
	return "contextValueSubroutine"
}
