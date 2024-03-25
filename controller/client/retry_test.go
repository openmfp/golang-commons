package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openmfp/golang-commons/controller/testSupport"
)

func TestRetry(t *testing.T) {
	o := &testSupport.TestApiObject{
		ObjectMeta: v1.ObjectMeta{Name: "test", Namespace: "test"},
	}
	c := testSupport.CreateFakeClient(t, o)

	t.Run("Retry status update", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Act
		err := RetryStatusUpdate(ctx, func(object client.Object) client.Object {
			return object
		}, o, c)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("Retry update", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Act
		err := RetryUpdate(ctx, func(object client.Object) client.Object {
			return object
		}, o, c)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("Retry update and fail", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		newObject := &testSupport.TestApiObject{
			ObjectMeta: v1.ObjectMeta{Name: "test1", Namespace: "test1"},
		}

		// Act
		err := RetryUpdate(ctx, func(object client.Object) client.Object {
			return object
		}, newObject, c)

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.IsNotFound(err))
	})

	t.Run("Retry Status update and fail", func(t *testing.T) {
		// Arrange
		ctx := context.Background()
		newObject := &testSupport.TestApiObject{
			ObjectMeta: v1.ObjectMeta{Name: "test1", Namespace: "test1"},
		}

		// Act
		err := RetryStatusUpdate(ctx, func(object client.Object) client.Object {
			return object
		}, newObject, c)

		// Assert
		assert.Error(t, err)
		assert.True(t, errors.IsNotFound(err))
	})
}
