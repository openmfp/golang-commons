package config_test

import (
	"context"
	"testing"
	"time"

	"github.com/openmfp/golang-commons/config"
	"github.com/stretchr/testify/assert"
)

func TestSetConfigInContext(t *testing.T) {
	ctx := context.Background()
	configStr := "test"
	ctx = config.SetConfigInContext(ctx, configStr)

	retrievedConfig := config.LoadConfigFromContext(ctx)
	assert.Equal(t, configStr, retrievedConfig)
}

func TestNewConfigFor(t *testing.T) {

	type test struct {
		config.CommonServiceConfig
		CustomFlag       string `mapstructure:"custom-flag"`
		CustomFlagInt    int    `mapstructure:"custom-flag-int"`
		CustomFlagBool   bool   `mapstructure:"custom-flag-bool"`
		CustomFlagStruct struct {
			CustomFlagDuration time.Duration `mapstructure:"custom-flag-duration"`
		} `mapstructure:",squash"`
		CustomFlagStruct2 struct {
			CustomFlagDuration time.Duration `mapstructure:"custom-flag-duration-2"`
		} `mapstructure:"le-flag"`
	}

	testStruct := test{}

	v, err := config.NewConfigFor(&testStruct)
	assert.NoError(t, err)

	err = v.Unmarshal(&testStruct)
	assert.NoError(t, err)

}
