package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetConfigInContext(t *testing.T) {
	ctx := context.Background()
	config := "test"
	ctx = SetConfigInContext(ctx, config)

	retrievedConfig := LoadConfigFromContext(ctx)
	assert.Equal(t, config, retrievedConfig)
}
