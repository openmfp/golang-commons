package lifecycle

import (
	"testing"
	"time"

	"github.com/openmfp/golang-commons/controller/testSupport"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openmfp/golang-commons/logger/testlogger"
)

func TestGetNextReconcilationTime(t *testing.T) {
	expectedEaliest := 12 * time.Hour
	expectedLatest := 24 * time.Hour

	actual := getNextReconcileTime()
	if actual < expectedEaliest || actual > expectedLatest {
		t.Errorf("Expected time between %v and %v, but got %v", expectedEaliest, expectedLatest, actual)
	}

	actual2 := getNextReconcileTime()
	if actual2 < expectedEaliest || actual2 > expectedLatest {
		t.Errorf("Expected time between %v and %v, but got %v", expectedEaliest, expectedLatest, actual)
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
	tl := testlogger.New()
	removeRefreshLabelIfExists(apiObject, tl.Logger)

	_, ok := apiObject.GetLabels()[SpreadReconcileRefreshLabel]

	messages, err := tl.GetLogMessages()
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Contains(t, messages[0].Message, "Removing refresh label")
}

func TestRemoveRefreshLabelFilledWithValue(t *testing.T) {
	apiObject := &testSupport.TestApiObject{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{SpreadReconcileRefreshLabel: "true"},
		},
	}
	tl := testlogger.New()
	removeRefreshLabelIfExists(apiObject, tl.Logger)

	_, ok := apiObject.GetLabels()[SpreadReconcileRefreshLabel]

	messages, err := tl.GetLogMessages()
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Contains(t, messages[0].Message, "Removing refresh label")
}

func TestRemoveRefreshLabelNoLabels(t *testing.T) {
	apiObject := &testSupport.TestApiObject{
		ObjectMeta: v1.ObjectMeta{},
	}
	tl := testlogger.New()
	removeRefreshLabelIfExists(apiObject, tl.Logger)

	_, ok := apiObject.GetLabels()[SpreadReconcileRefreshLabel]

	messages, err := tl.GetLogMessages()
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Contains(t, messages[0].Message, "Refresh label not found")
}
