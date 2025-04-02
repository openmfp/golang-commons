package config

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/openmfp/golang-commons/context/keys"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func SetConfigInContext(ctx context.Context, config any) context.Context {
	return context.WithValue(ctx, keys.ConfigCtxKey, config)
}

func LoadConfigFromContext(ctx context.Context) any {
	return ctx.Value(keys.ConfigCtxKey)
}

type CommonServiceConfig struct {
	DebugLabelValue         string `mapstructure:"debug-label-value"`
	MaxConcurrentReconciles int    `mapstructure:"max-concurrent-reconciles"`
	Environment             string `mapstructure:"environment"`
	Region                  string `mapstructure:"region"`
	Kubeconfig              string `mapstructure:"kubeconfig"`
	Image                   struct {
		Name string `mapstructure:"image-name"`
		Tag  string `mapstructure:"image-tag"`
	} `mapstructure:",squash"`
	Log struct {
		Level string `mapstructure:"log-level"`

		NoJson bool `mapstructure:"no-json"`
	} `mapstructure:",squash"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown-timeout"`
	Probes          struct {
		BindAddress string `mapstructure:"probes-bind-address"`
	} `mapstructure:",squash"`
	LeaderElection struct {
		Enabled bool `mapstructure:"leader-election-enabled"`
	} `mapstructure:",squash"`
	Sentry struct {
		Dsn string `mapstructure:"sentry-dsn"`
	} `mapstructure:",squash"`
}

func CommonFlags() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("common", pflag.ContinueOnError)

	flagSet.String("debug-label-value", "", "Debug label value")
	flagSet.Int("max-concurrent-reconciles", 1, "Max concurrent reconciles")
	flagSet.String("environment", "local", "Environment")
	flagSet.String("region", "local", "Region")
	flagSet.String("image-name", "", "Image name")
	flagSet.String("image-tag", "latest", "Image tag")
	flagSet.String("log-level", "info", "Log level")
	flagSet.Bool("log-no-json", false, "Log in JSON format")
	flagSet.Duration("shutdown-timeout", 1, "Shutdown timeout")
	flagSet.String("probes-bind-address", ":8090", "Probes bind address")
	flagSet.Bool("leader-election-enabled", false, "Enable leader election")
	flagSet.String("sentry-dsn", "", "Sentry DSN")

	return flagSet
}

// generateFlagSet generates a pflag.FlagSet from a struct based on its `mapstructure` tags.
func generateFlagSet(config any) *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("generated", pflag.ContinueOnError)
	traverseStruct(reflect.ValueOf(config), flagSet, "")
	return flagSet
}

// traverseStruct recursively traverses a struct and adds flags to the FlagSet.
func traverseStruct(value reflect.Value, flagSet *pflag.FlagSet, prefix string) {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return
	}

	typ := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := value.Field(i)

		// Get the `mapstructure` tag
		tag := field.Tag.Get("mapstructure")
		if tag == "" || tag == "-" {
			continue
		}

		// Handle nested structs
		if fieldValue.Kind() == reflect.Struct {
			if tag == ",squash" {
				traverseStruct(fieldValue, flagSet, "")
			} else {
				traverseStruct(fieldValue, flagSet, prefix+tag+".")
			}
			continue
		}

		// Add flags based on the field type
		switch fieldValue.Kind() {
		case reflect.String:
			flagSet.String(prefix+tag, "", fmt.Sprintf("Set the %s", tag))
		case reflect.Int, reflect.Int64:
			if fieldValue.Type() == reflect.TypeOf(time.Duration(0)) {

				flagSet.Duration(prefix+tag, 0, fmt.Sprintf("Set the %s", tag))
			} else {
				flagSet.Int(prefix+tag, 0, fmt.Sprintf("Set the %s", tag))
			}
		case reflect.Bool:
			flagSet.Bool(prefix+tag, false, fmt.Sprintf("Set the %s", tag))
		}
	}
}

func NewConfigFor(serviceConfig any) *viper.Viper {
	v := viper.NewWithOptions(
		viper.EnvKeyReplacer(strings.NewReplacer("-", "_")),
	)

	v.AutomaticEnv()

	v.BindPFlags(CommonFlags())
	v.BindPFlags(generateFlagSet(serviceConfig))

	return v
}
