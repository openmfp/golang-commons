package testSupport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTestApiObject_DeepCopy(t *testing.T) {
	t.Run("DeepCopyObject", func(t *testing.T) {
		// Arrange
		instance := &TestApiObject{}

		// Act
		c := instance.DeepCopyObject()

		// Assert
		assert.Equal(t, instance, c)
	})

	t.Run("DeepCopyObject with nil", func(t *testing.T) {
		// Arrange
		var instance *TestApiObject

		// Act
		c := instance.DeepCopyObject()

		// Assert
		assert.Nil(t, c)
	})

	t.Run("DeepCopy", func(t *testing.T) {
		// Arrange
		var instance *TestApiObject

		// Act
		c := instance.DeepCopy()

		// Assert
		assert.Equal(t, instance, c)
	})
}
