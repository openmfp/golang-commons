package client

import (
	"context"
	"testing"

	"github.com/openmfp/golang-commons/controller/testSupport"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
}
