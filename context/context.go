// Package context implements a safe way to create a context which will have a timeout after closing the context
//
// The basic usage is using the StartContext function with a direct defer call to shutdown afterwards.
// There is a Timeout set in the context so TimeoutFromContext will return the time this context will wait for subroutines to stop.
package context

import (
	"context"
	"time"

	"github.com/openmfp/golang-commons/config"
	"github.com/openmfp/golang-commons/logger"
)

// Can be used in StartContext to have a sane default timeout
const DefaultShutdownTimeout = 3 * time.Second

type ShutdownTimeoutKey struct{}

// Creates a new context and returns context, cancel and shutdown function. It should be directly followed by a defer call to shutdown.
func StartContext(log *logger.Logger, cfg any, timeout time.Duration) (ctx context.Context, cancel context.CancelCauseFunc, shutdown func()) {
	parentCtx := context.WithValue(context.Background(), ShutdownTimeoutKey{}, timeout)
	ctxWithCause, cancel := context.WithCancelCause(parentCtx)
	ctx, c := NotifyShutdownContext(ctxWithCause)
	ctx = logger.SetLoggerInContext(ctx, log)
	ctx = config.SetConfigInContext(ctx, cfg)

	return ctx, cancel, func() {
		Recover(log)
		c()
		<-time.After(timeout)
	}
}

// Returns the wrapped timeout inside the context and will return a default timeout of 3 seconds when the context does not exist or doesn't wrap a timeout.
func TimeoutFromContext(ctx context.Context) time.Duration {
	if ctx == nil {
		return DefaultShutdownTimeout
	}
	timeout, ok := ctx.Value(ShutdownTimeoutKey{}).(time.Duration)
	if !ok {
		return DefaultShutdownTimeout
	}
	return timeout
}
