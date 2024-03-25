package errors

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

func TestRetry(t *testing.T) {

	t.Run("Is retryable with an internal error", func(t *testing.T) {
		// Arrange
		e := k8sErrors.NewServiceUnavailable("na")

		// Act
		retriable, result := IsRetriable(e)

		// Assert
		assert.True(t, retriable)
		assert.Equal(t, time.Duration(0), result.RequeueAfter)
	})

	t.Run("Is retryable with a unknown error", func(t *testing.T) {
		// Arrange
		e := fmt.Errorf("oh nose")

		// Act
		retriable, result := IsRetriable(e)

		// Assert
		assert.False(t, retriable)
		assert.Equal(t, time.Duration(0), result.RequeueAfter)
	})

	t.Run("Is retryable with a clientDelay", func(t *testing.T) {
		// Arrange
		e := k8sErrors.NewTimeoutError("oh nose", 5)

		// Act
		retriable, result := IsRetriable(e)

		// Assert
		assert.True(t, retriable)
		assert.NotEqualf(t, time.Duration(0), result.RequeueAfter, "Expected requeueAfter to be set, but got %v", result.RequeueAfter)
	})
}
