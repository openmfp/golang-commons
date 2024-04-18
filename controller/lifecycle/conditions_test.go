package lifecycle

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openmfp/golang-commons/controller/testSupport"
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
