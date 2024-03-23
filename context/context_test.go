package context

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/openmfp/golang-commons/logger"
)

type TestConfig struct {
	Key string
}

func TestStartContext(t *testing.T) {
	t.Parallel()

	log, _ := logger.NewTestLogger()
	cfg := TestConfig{Key: "value"}

	ctx, cancel, shutdown := StartContext(log, cfg, 3*time.Second)
	defer shutdown()

	assert.NotNil(t, ctx)
	assert.NotNil(t, cancel)
}

func TestDefaultTimeout(t *testing.T) {
	t.Parallel()

	timeout := TimeoutFromContext(nil)
	assert.Equal(t, DefaultShutdownTimeout, timeout)

	timeoutCtx := TimeoutFromContext(context.Background())
	assert.Equal(t, DefaultShutdownTimeout, timeoutCtx)
}

func TestTimeoutFromContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	ctx = context.WithValue(ctx, ShutdownTimeoutKey{}, 5*time.Second)

	timeout := TimeoutFromContext(ctx)
	assert.Equal(t, 5*time.Second, timeout)
}
