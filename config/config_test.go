package config_test

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/openmfp/golang-commons/config"
)

func TestSetConfigInContext(t *testing.T) {
	ctx := context.Background()
	configStr := "test"
	ctx = config.SetConfigInContext(ctx, configStr)

	retrievedConfig := config.LoadConfigFromContext(ctx)
	assert.Equal(t, configStr, retrievedConfig)
}

func TestBindConfigToFlags(t *testing.T) {

	type test struct {
		config.CommonServiceConfig
		CustomFlag          string `mapstructure:"custom-flag" default:"abc" description:"This is a custom flag"`
		CustomFlagInt       int    `mapstructure:"custom-flag-int" default:"123" description:"This is a custom flag with int value"`
		CustomFlagBool      bool   `mapstructure:"custom-flag-bool" default:"true" description:"This is a custom flag with bool value"`
		CustomFlagNoDefault string `mapstructure:"custom-flag-no-default" `
		CustomFlagStruct    struct {
			CustomFlagDuration time.Duration `mapstructure:"custom-flag-duration" default:"1m" description:"This is a custom flag with duration value"`
			SubCustomFlag      string        `mapstructure:"sub-custom-flag" default:"subabc" description:"This is a sub custom flag"`
		} `mapstructure:",squash"`
		CustomFlagStruct2 struct {
			CustomFlagDuration time.Duration `mapstructure:"custom-flag-duration-2"`
		} `mapstructure:"le-strFlag"`
	}

	testStruct := test{}

	v := viper.New()

	cmd := &cobra.Command{}
	_ = config.BindConfigToFlags(v, cmd, &testStruct) // assuming this binds flags

	strFlag := cmd.Flags().Lookup("custom-flag")
	assert.NotNil(t, strFlag)
	assert.Equal(t, "This is a custom flag", strFlag.Usage)
	assert.Equal(t, "abc", strFlag.DefValue)

	subStrFlag := cmd.Flags().Lookup("sub-custom-flag")
	assert.NotNil(t, subStrFlag)
	assert.Equal(t, "This is a sub custom flag", subStrFlag.Usage)
	assert.Equal(t, "subabc", subStrFlag.DefValue)

	strNoDefaultFlag := cmd.Flags().Lookup("custom-flag-no-default")
	assert.NotNil(t, strNoDefaultFlag)
	assert.Equal(t, "Set the custom-flag-no-default", strNoDefaultFlag.Usage)
	assert.Equal(t, "", strNoDefaultFlag.DefValue)

	intFlag := cmd.Flags().Lookup("custom-flag-int")
	assert.NotNil(t, intFlag)
	assert.Equal(t, "This is a custom flag with int value", intFlag.Usage)
	assert.Equal(t, "123", intFlag.DefValue)

	boolFlag := cmd.Flags().Lookup("custom-flag-bool")
	assert.NotNil(t, boolFlag)
	assert.Equal(t, "This is a custom flag with bool value", boolFlag.Usage)
	assert.Equal(t, "true", boolFlag.DefValue)

	durationFlag := cmd.Flags().Lookup("custom-flag-duration")
	assert.NotNil(t, durationFlag)
	assert.Equal(t, "This is a custom flag with duration value", durationFlag.Usage)
	assert.Equal(t, "1m0s", durationFlag.DefValue)
}

func TestBindConfigToFlagsWrongTypeInt(t *testing.T) {
	type test struct {
		CustomFlagInt int `mapstructure:"custom-flag-int" default:"abc" description:"This is a custom flag with int value"`
	}

	testStruct := test{}

	v := viper.New()

	cmd := &cobra.Command{}
	err := config.BindConfigToFlags(v, cmd, &testStruct) // assuming this binds flags
	assert.Error(t, err)
}

func TestBindConfigToFlagsWrongTypeDuration(t *testing.T) {
	type test struct {
		CustomFlagInt time.Duration `mapstructure:"custom-flag" default:"abc"`
	}

	testStruct := test{}

	v := viper.New()

	cmd := &cobra.Command{}
	err := config.BindConfigToFlags(v, cmd, &testStruct) // assuming this binds flags
	assert.Error(t, err)
}

func TestBindConfigToFlagsWrongTypeBool(t *testing.T) {
	type test struct {
		CustomFlagInt bool `mapstructure:"custom-flag" default:"abc"`
	}

	testStruct := test{}

	v := viper.New()

	cmd := &cobra.Command{}
	err := config.BindConfigToFlags(v, cmd, &testStruct) // assuming this binds flags
	assert.Error(t, err)
}

func TestBindConfigToFlagsEmptyDefaultBool(t *testing.T) {
	type test struct {
		CustomFlagInt bool `mapstructure:"custom-flag" default:""`
	}

	testStruct := test{}

	v := viper.New()

	cmd := &cobra.Command{}
	err := config.BindConfigToFlags(v, cmd, &testStruct) // assuming this binds flags
	assert.NoError(t, err)

	boolFlag := cmd.Flags().Lookup("custom-flag")
	assert.NotNil(t, boolFlag)
	assert.Equal(t, "Set the custom-flag", boolFlag.Usage)
	assert.Equal(t, "false", boolFlag.DefValue)
}

func TestBindConfigToFlagsWrongType(t *testing.T) {
	type test struct {
		CustomFlagInt byte `mapstructure:"custom-flag" default:"abc"`
	}

	testStruct := test{}

	v := viper.New()

	cmd := &cobra.Command{}
	err := config.BindConfigToFlags(v, cmd, &testStruct) // assuming this binds flags
	assert.Error(t, err)
}

func TestBindConfigToFlagsWrongTypeNoStruct(t *testing.T) {
	testStruct := ""

	v := viper.New()

	cmd := &cobra.Command{}
	err := config.BindConfigToFlags(v, cmd, &testStruct) // assuming this binds flags
	assert.Error(t, err)
}

func TestNewDefaultConfig(t *testing.T) {
	v, _, err := config.NewDefaultConfig(&cobra.Command{})
	assert.NoError(t, err)

	err = v.Unmarshal(&config.CommonServiceConfig{})
	assert.NoError(t, err)
}
