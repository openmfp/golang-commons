package logger

import (
	"context"
	"os"
	"testing"

	"github.com/openmfp/golang-commons/context/keys"
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
	logger, err := New(DefaultConfig())
	assert.NoError(t, err)
	ctx = SetLoggerInContext(ctx, logger)

	retrievedLogger := LoadLoggerFromContext(ctx)
	assert.NotNil(t, retrievedLogger)
}

func TestTestLoggerInContextFallback(t *testing.T) {
	ctx := context.Background()
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

func TestNewRequestLoggerFromZeroLogCtxWithValue(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, keys.RequestIdCtxKey, "test")
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

func TestNewWithUnknownLogLevel(t *testing.T) {
	logger, err := New(Config{Level: "unknown"})
	assert.Nil(t, logger)
	assert.Error(t, err)
}

func TestComponentLogger(t *testing.T) {
	cfg := DefaultConfig()
	cfg.NoJSON = true
	logger, _ := New(cfg)
	componentLogger := logger.ComponentLogger("my-component")
	assert.NotNil(t, componentLogger)
	componentLogger.Level(1).Debug().Msg("test")
	componentLogger.Logr().Info("test")
}

func TestChildLoggerWithAttributes(t *testing.T) {
	logger, _ := New(DefaultConfig())
	_, err := logger.ChildLoggerWithAttributes("key", "value")
	assert.NoError(t, err)
}

func TestMustChildLoggerWithAttributes(t *testing.T) {
	logger, _ := New(DefaultConfig())
	loggerAttributes := logger.MustChildLoggerWithAttributes("key", "value")
	assert.NotNil(t, loggerAttributes)
}
