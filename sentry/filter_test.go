package sentry

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	openmfperrors "github.com/openmfp/golang-commons/errors"
)

func TestIsTerminatingNSError(t *testing.T) {
	gvr := schema.GroupResource{
		Group:    "a.api.group",
		Resource: "aResource",
	}

	conflictError := errors.NewConflict(gvr, "aName", fmt.Errorf("aMessage"))
	conflictError.ErrStatus.Details.Causes = []metav1.StatusCause{{Type: v1.NamespaceTerminatingCause}}
	isTerminating := isTerminatingNSError(conflictError)
	assert.True(t, isTerminating)
}

func TestShouldBeProcessedNegative(t *testing.T) {
	gvr := schema.GroupResource{
		Group:    "a.api.group",
		Resource: "aResource",
	}

	conflictError := errors.NewConflict(gvr, "aName", fmt.Errorf("aMessage"))
	conflictError.ErrStatus.Details.Causes = []metav1.StatusCause{{Type: v1.NamespaceTerminatingCause}}
	shouldBeProcessed := ShouldBeProcessed(conflictError)
	assert.False(t, shouldBeProcessed)
}

func TestShouldBeProcessedPositive(t *testing.T) {
	gvr := schema.GroupResource{
		Group:    "a.api.group",
		Resource: "aResource",
	}

	conflictError := errors.NewConflict(gvr, "aName", fmt.Errorf("aMessage"))
	shouldBeProcessed := ShouldBeProcessed(conflictError)
	assert.True(t, shouldBeProcessed)
}

func TestShouldBeProcessedSentryPositive(t *testing.T) {
	err := openmfperrors.New("test error")
	sentryError := SentryError(err)
	shouldBeProcessed := ShouldBeProcessed(sentryError)
	assert.True(t, shouldBeProcessed)
}
