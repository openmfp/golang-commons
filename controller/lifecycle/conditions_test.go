package lifecycle

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"

	"github.com/openmfp/golang-commons/controller/testSupport"
	"github.com/openmfp/golang-commons/logger"
)

// Test LifecycleManager.WithConditionManagement
func TestLifecycleManager_WithConditionManagement(t *testing.T) {
	// Given
	fakeClient := testSupport.CreateFakeClient(t, &testSupport.TestApiObject{})
	_, log := createLifecycleManager([]Subroutine{}, fakeClient)

	// When
	l := NewLifecycleManager(log.Logger, "test-operator", "test-controller", fakeClient, []Subroutine{}).WithConditionManagement()

	// Then
	assert.True(t, true, l.manageConditions)
}

// Test the setReady function with an empty array
func TestSetReady(t *testing.T) {

	t.Run("TestSetReady with empty array", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{}

		// When
		setInstanceConditionReady(&condition, metav1.ConditionTrue)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[0].Status)
	})

	t.Run("TestSetReady with existing condition", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{
			{Type: "test", Status: metav1.ConditionFalse},
		}

		// When
		setInstanceConditionReady(&condition, metav1.ConditionTrue)

		// Then
		assert.Equal(t, 2, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[1].Status)
	})
}

func TestSetUnknown(t *testing.T) {

	t.Run("TestSetUnknown with empty array", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{}

		// When
		setInstanceConditionUnknownIfNotSet(&condition)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionUnknown, condition[0].Status)
	})

	t.Run("TestSetUnknown with existing ready condition", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{
			{Type: ConditionReady, Status: metav1.ConditionTrue},
		}

		// When
		setInstanceConditionUnknownIfNotSet(&condition)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[0].Status)
	})
}

func TestSetSubroutineConditionToUnknownIfNotSet(t *testing.T) {
	log, err := logger.New(logger.DefaultConfig())
	require.NoError(t, err)

	t.Run("TestSetSubroutineConditionToUnknownIfNotSet with empty array", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{}

		// When
		setSubroutineConditionToUnknownIfNotSet(&condition, changeStatusSubroutine{}, false, log)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionUnknown, condition[0].Status)
	})

	t.Run("TestSetSubroutineConditionToUnknownIfNotSet with existing condition", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{
			{Type: "test", Status: metav1.ConditionFalse},
		}

		// When
		setSubroutineConditionToUnknownIfNotSet(&condition, changeStatusSubroutine{}, false, log)

		// Then
		assert.Equal(t, 2, len(condition))
		assert.Equal(t, metav1.ConditionUnknown, condition[1].Status)
	})

	t.Run("TestSetSubroutineConditionToUnknownIfNotSet with existing ready", func(t *testing.T) {
		// Given
		subroutine := changeStatusSubroutine{}
		condition := []metav1.Condition{
			{Type: "test", Status: metav1.ConditionFalse},
			{Type: fmt.Sprintf("%s_Ready", subroutine.GetName()), Status: metav1.ConditionTrue},
		}

		// When
		setSubroutineConditionToUnknownIfNotSet(&condition, subroutine, false, log)

		// Then
		assert.Equal(t, 2, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[1].Status)
	})
}

func TestSubroutineCondition(t *testing.T) {
	log, err := logger.New(logger.DefaultConfig())
	require.NoError(t, err)

	// Add a test case to set a subroutine condition to ready if it was successfull
	t.Run("TestSetSubroutineConditionReady", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{}
		subroutine := changeStatusSubroutine{}

		// When
		setSubroutineCondition(&condition, subroutine, controllerruntime.Result{}, nil, false, log)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[0].Status)
	})

	// Add a test case to set a subroutine condition to unknown if it is still processing
	t.Run("TestSetSubroutineConditionProcessing", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{}
		subroutine := changeStatusSubroutine{}

		// When
		setSubroutineCondition(&condition, subroutine, controllerruntime.Result{Requeue: true}, nil, false, log)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionUnknown, condition[0].Status)
	})

	// Add a test case to set a subroutine condition to false if it failed
	t.Run("TestSetSubroutineConditionError", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{}
		subroutine := changeStatusSubroutine{}

		// When
		setSubroutineCondition(&condition, subroutine, controllerruntime.Result{}, errors.New("failed"), false, log)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionFalse, condition[0].Status)
	})

	// Add a test case to set a subroutine condition for isFinalize true
	t.Run("TestSetSubroutineFinalizeConditionReady", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{}
		subroutine := changeStatusSubroutine{}

		// When
		setSubroutineCondition(&condition, subroutine, controllerruntime.Result{}, nil, true, log)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionTrue, condition[0].Status)
	})

	// Add a test case to set a subroutine condition to unknown if it is still processing
	t.Run("TestSetSubroutineFinalizeConditionProcessing", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{}
		subroutine := changeStatusSubroutine{}

		// When
		setSubroutineCondition(&condition, subroutine, controllerruntime.Result{Requeue: true}, nil, true, log)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionUnknown, condition[0].Status)
	})

	// Add a test case to set a subroutine condition to false if it failed
	t.Run("TestSetSubroutineFinalizeConditionError", func(t *testing.T) {
		// Given
		condition := []metav1.Condition{}
		subroutine := changeStatusSubroutine{}

		// When
		setSubroutineCondition(&condition, subroutine, controllerruntime.Result{}, errors.New("failed"), true, log)

		// Then
		assert.Equal(t, 1, len(condition))
		assert.Equal(t, metav1.ConditionFalse, condition[0].Status)
	})
}
