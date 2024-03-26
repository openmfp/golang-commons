package context

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	testlogger "github.com/openmfp/golang-commons/logger/testlogger"
)

func TestRecover(t *testing.T) {
	t.Parallel()
	t.Run("should recover from panic and log", func(t *testing.T) {

		log := testlogger.New()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer Recover(log.Logger)
			panic("test panic")
		}()
		wg.Wait()

		logMessages, err := log.GetLogMessages()
		assert.NoError(t, err)
		require.Len(t, logMessages, 1)
		assert.Equal(t, "recovered panic", logMessages[0].Message)
	})

	t.Run("should recover from panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			defer Recover(nil)
			panic("test panic")
		})
	})
}
