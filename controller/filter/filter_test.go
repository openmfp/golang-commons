package filter

import (
	"testing"

	"github.com/openmfp/golang-commons/controller/testSupport"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

func TestFilter(t *testing.T) {
	predicate := DebugResourcesBehaviourPredicate("test")

	t.Run("Filter out test", func(t *testing.T) {
		// Arrange
		object := &testSupport.TestApiObject{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{DebugLabel: "test"}}}
		c := event.CreateEvent{Object: object}
		u := event.UpdateEvent{ObjectOld: object, ObjectNew: object}
		d := event.DeleteEvent{Object: object}
		g := event.GenericEvent{Object: object}

		// Act
		val := predicate.Create(c)
		val = predicate.Update(u) || val
		val = predicate.Delete(d) || val
		val = predicate.Generic(g) || val

		// Assert
		assert.True(t, val)
	})

	t.Run("Accept test", func(t *testing.T) {
		// Arrange
		object := &testSupport.TestApiObject{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{}}}
		c := event.CreateEvent{Object: object}
		u := event.UpdateEvent{ObjectOld: object, ObjectNew: object}
		d := event.DeleteEvent{Object: object}
		g := event.GenericEvent{Object: object}

		// Act
		val := predicate.Create(c)
		val = predicate.Update(u) && val
		val = predicate.Delete(d) && val
		val = predicate.Generic(g) && val

		// Assert
		assert.False(t, val)
	})
}
