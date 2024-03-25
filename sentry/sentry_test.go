package sentry

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	err := Start(context.Background(), "", "", "", "", "")
	assert.NoError(t, err)
}

func TestCaptureError(t *testing.T) {
	assert.NotPanics(t, func() {
		err := fmt.Errorf("test error")
		CaptureError(err, nil)
	})
}

func TestCaptureSentryError(t *testing.T) {

	assert.NotPanics(t, func() {
		err := SentryError(fmt.Errorf("test error"))
		CaptureSentryError(err, nil)
	})
}
