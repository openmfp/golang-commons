package context

import (
	"runtime/debug"
	"time"

	"github.com/getsentry/sentry-go"

	"github.com/openmfp/golang-commons/logger"
)

// Recover can be used as deferred function to catch panics
// This function is used in the context of context creation. Its contained in the context package to avoid circular dependencies with sentry package
func Recover(log *logger.Logger) {
	if log == nil {
		log = logger.StdLogger
	}

	if err := recover(); err != nil {
		log.Error().Interface("panic", err).Interface("stack", debug.Stack()).Msg("recovered panic")
		sentry.CurrentHub().Recover(err)
		sentry.Flush(time.Second * 5)
	}
}
