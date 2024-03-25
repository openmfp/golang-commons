package context

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecover(t *testing.T) {
	t.Run("should recover from panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			defer Recover(nil)
			panic("test panic")
		})
	})
}
