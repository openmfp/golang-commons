package lifecycle

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openmfp/golang-commons/controller/testSupport"
	"github.com/openmfp/golang-commons/logger/testlogger"
)

func TestGetNextReconcilationTime(t *testing.T) {
	expectedEarliest := 12 * time.Hour
	expectedLatest := 24 * time.Hour

	actual := getNextReconcileTime(defaultMaxReconcileDuration)
	if actual < expectedEarliest || actual > expectedLatest {
		t.Errorf("Expected time between %v and %v, but got %v", expectedEarliest, expectedLatest, actual)
	}

	actual2 := getNextReconcileTime(defaultMaxReconcileDuration)
	if actual2 < expectedEarliest || actual2 > expectedLatest {
		t.Errorf("Expected time between %v and %v, but got %v", expectedEarliest, expectedLatest, actual)
	}

	if actual == actual2 {
		t.Errorf("Expected different values, but got same")
	}
}

func TestOnNextReconcile(t *testing.T) {
	nextReconcile := time.Now().Add(10 * time.Minute)
	instanceStatusObj := testSupport.TestStatus{
		NextReconcileTime: v1.NewTime(nextReconcile),
	}
	apiObject := &implementingSpreadReconciles{testSupport.TestApiObject{Status: instanceStatusObj}}
	tl := testlogger.New()

	requeueAfter, err := onNextReconcile(apiObject, tl.Logger)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	timeTill := time.Until(nextReconcile.UTC())
	if time.Duration(requeueAfter.RequeueAfter.Seconds()) != time.Duration(timeTill.Seconds()) {
		t.Errorf("Expected requeueAfter to be %v, but got %v", timeTill, requeueAfter.RequeueAfter)
	}
	messages, err := tl.GetLogMessages()
	assert.NoError(t, err)
	assert.Contains(t, messages[0].Message, "no processing needed")
}

type testInstance struct {
	mock.Mock
	*implementingSpreadReconciles
}

func (t *testInstance) GenerateNextReconcileTime() time.Duration {
	args := t.Called()
	return args.Get(0).(time.Duration)
}

func TestGenerateNextReconcileTimer(t *testing.T) {
	instance := &testInstance{
		implementingSpreadReconciles: &implementingSpreadReconciles{testSupport.TestApiObject{}},
	}

	instance.On("GenerateNextReconcileTime").Return(10 * time.Minute)

	setNextReconcileTime(instance, testlogger.New().Logger)

	assert.True(t, instance.AssertCalled(t, "GenerateNextReconcileTime"))
}

func TestUpdateObservedGeneration(t *testing.T) {
	instanceStatusObj := testSupport.TestStatus{
		ObservedGeneration: 0,
	}
	apiObject := &implementingSpreadReconciles{testSupport.TestApiObject{
		Status: instanceStatusObj,
		ObjectMeta: v1.ObjectMeta{
			Generation: 1,
		},
	},
	}
	tl := testlogger.New()
	updateObservedGeneration(apiObject, tl.Logger)

	assert.Equal(t, apiObject.GetObservedGeneration(), apiObject.GetGeneration())
	messages, err := tl.GetLogMessages()
	assert.NoError(t, err)
	assert.Contains(t, messages[0].Message, "Updating observed generation")
}

func TestRemoveRefreshLabel(t *testing.T) {
	apiObject := &testSupport.TestApiObject{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{SpreadReconcileRefreshLabel: ""},
		},
	}
	removeRefreshLabelIfExists(apiObject)

	_, ok := apiObject.GetLabels()[SpreadReconcileRefreshLabel]
	assert.False(t, ok)
}

func TestRemoveRefreshLabelFilledWithValue(t *testing.T) {
	apiObject := &testSupport.TestApiObject{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{SpreadReconcileRefreshLabel: "true"},
		},
	}
	removeRefreshLabelIfExists(apiObject)

	_, ok := apiObject.GetLabels()[SpreadReconcileRefreshLabel]

	assert.False(t, ok)
}

func TestRemoveRefreshLabelNoLabels(t *testing.T) {
	apiObject := &testSupport.TestApiObject{
		ObjectMeta: v1.ObjectMeta{},
	}
	removeRefreshLabelIfExists(apiObject)

	_, ok := apiObject.GetLabels()[SpreadReconcileRefreshLabel]

	assert.False(t, ok)
}
