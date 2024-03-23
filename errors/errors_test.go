package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewErrorWithStackTrace(t *testing.T) {
	err := New("oops")
	assert.Equal(t, "oops", err.Error())
	assert.Len(t, getStackTraces(err), 1)
}

func TestNewOperatorErrorWithStackTrace(t *testing.T) {
	err := NewOperatorError(New("oops"), false, false)
	assert.Equal(t, "oops", err.Err().Error())
	assert.Len(t, getStackTraces(err.Err()), 1)
	assert.False(t, err.Retry())
	assert.False(t, err.Sentry())
}

func TestWithStackTraceExistingError(t *testing.T) {
	cause := errors.New("failed")
	err := WithStack(cause)

	if err, ok := err.(StackTracer); ok {
		fmt.Printf("%+v\n", err.StackTrace())
	}

	assert.Equal(t, "failed", err.Error())
	assert.Len(t, getStackTraces(err), 1)
}

func TestWithStackTraceExistingErrorAndStack(t *testing.T) {
	cause := New("failed")
	err := WithStack(cause)

	if err, ok := err.(StackTracer); ok {
		fmt.Printf("%+v\n", err.StackTrace())
	}

	assert.Equal(t, "failed", err.Error())
	// 2 Stacks the one from the first error and the added one from WithStack
	assert.Len(t, getStackTraces(err), 2)
}

func TestEnsureStackWithExistingErrorAndStack(t *testing.T) {
	cause := New("failed")
	err := EnsureStack(cause)

	if err, ok := err.(StackTracer); ok {
		fmt.Printf("%+v\n", err.StackTrace())
	}

	assert.Equal(t, "failed", err.Error())
	assert.Len(t, getStackTraces(err), 1)
}

func TestWithStackTraceExistingErrorWithDifferentStack(t *testing.T) {
	cause := subFunc()
	err := WithStack(cause)

	if err, ok := err.(StackTracer); ok {
		fmt.Printf("%+v\n", err.StackTrace())
	}

	assert.Equal(t, "failed", err.Error())
	assert.Len(t, getStackTraces(err), 2)
}

func subFunc() error {
	cause := New("failed")
	return cause
}

func TestWrapCauseWithNoStackTrace(t *testing.T) {
	err := Wrap(fmt.Errorf("cause"), "description")
	assert.Equal(t, "description: cause", err.Error())

	cause := Cause(err)
	assert.Equal(t, "cause", cause.Error())
	assert.Len(t, getStackTraces(err), 1)
}

func TestWrapCauseWithStackTrace(t *testing.T) {
	additionalStack(t)
}

func additionalStack(t *testing.T) {
	errChan := make(chan error)
	go func() {
		err := New("unrelated") // created with a stack trace
		errChan <- err
	}()

	receivedErr := <-errChan
	err := Wrap(receivedErr, "helpful description")
	assert.Equal(t, "helpful description: unrelated", err.Error())
	assert.Len(t, getStackTraces(err), 2)
}

func TestWrapFrameFromOUrCurrentMethod(t *testing.T) {
	err := Wrap(New("related"), "helpful description")
	assert.Equal(t, "helpful description: related", err.Error())
	assert.Len(t, getStackTraces(err), 1)
}

func getStackTraces(err error) []StackTrace {
	traces := []StackTrace{}
	if err, ok := err.(StackTracer); ok {
		traces = append(traces, err.StackTrace())
	}

	if err := Unwrap(err); err != nil {
		traces = append(traces, getStackTraces(err)...)
	}

	return traces
}
