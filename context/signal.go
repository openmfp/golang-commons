package context

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
)

type signalCtx struct {
	context.Context

	cancel  context.CancelCauseFunc
	signals []os.Signal
	ch      chan os.Signal
}

func (c *signalCtx) stop() {
	c.cancel(nil)
	signal.Stop(c.ch)
}

var ErrShutdown = errors.New("shutdown")

// NotifyShutdownContext returns a copy of the parent context that is marked done
// (its Done channel is closed) when one of the expected signals arrives,
// when the returned stop function is called, or when the parent context's
// Done channel is closed, whichever happens first.
func NotifyShutdownContext(parent context.Context) (ctx context.Context, stop context.CancelFunc) {
	ctx, cancel := context.WithCancelCause(parent)
	signals := []os.Signal{syscall.SIGKILL, syscall.SIGTERM, syscall.SIGINT}
	c := &signalCtx{
		Context: ctx,
		cancel:  cancel,
		signals: signals,
	}
	c.ch = make(chan os.Signal, 1)
	signal.Notify(c.ch, c.signals...)
	if ctx.Err() == nil {
		go func() {
			select {
			case <-c.ch:
				c.cancel(ErrShutdown)
			case <-c.Done():
			}
		}()
	}
	return c, c.stop
}
