package logger

import (
	"context"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestLoggerInContext(t *testing.T) {
	ctx := context.Background()
	logger, _ := New(DefaultConfig())
	ctx = SetLoggerInContext(ctx, logger)

	retrievedLogger := LoadLoggerFromContext(ctx)
	assert.NotNil(t, retrievedLogger)
}

func TestTestLoggerInContext(t *testing.T) {
	ctx := context.Background()
	logger, _ := NewTestLogger()
	ctx = SetLoggerInContext(ctx, logger)

	retrievedLogger := LoadLoggerFromContext(ctx)
	assert.NotNil(t, retrievedLogger)
}

func TestNewFromZeroLog(t *testing.T) {
	logger := NewFromZerolog(zerolog.New(os.Stdout))
	assert.NotNil(t, logger)
}

func TestNewRequestLoggerFromZeroLog(t *testing.T) {
	ctx := context.Background()
	logger := NewRequestLoggerFromZerolog(ctx, zerolog.New(os.Stdout))
	assert.NotNil(t, logger)
}

func TestNewChildLoggerRequestLoggerFromZeroLog(t *testing.T) {
	ctx := context.Background()
	logger := NewRequestLoggerFromZerolog(ctx, zerolog.New(os.Stdout))
	assert.NotNil(t, logger)

	childLogger := logger.ChildLogger("child", "my-child")
	assert.NotNil(t, childLogger)
}
