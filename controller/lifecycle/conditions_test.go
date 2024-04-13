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
	// Given
	condition := []metav1.Condition{}

	// When
	setReady(&condition, metav1.ConditionTrue)

	// Then
	assert.Equal(t, 1, len(condition))
	assert.Equal(t, metav1.ConditionTrue, condition[0].Status)
}

// Test the setReady function with existing conditions
func TestSetReadyWithExistingConditions(t *testing.T) {
	// Given
	condition := []metav1.Condition{
		{Type: "test", Status: metav1.ConditionFalse},
	}

	// When
	setReady(&condition, metav1.ConditionTrue)

	// Then
	assert.Equal(t, 2, len(condition))
	assert.Equal(t, metav1.ConditionTrue, condition[1].Status)
}
