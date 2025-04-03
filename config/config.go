package config

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/openmfp/golang-commons/context/keys"
	"github.com/spf13/cobra"
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

	Image struct {
		Name string `mapstructure:"image-name"`
		Tag  string `mapstructure:"image-tag"`
	} `mapstructure:",squash"`

	Log struct {
		Level string `mapstructure:"log-level"`

		NoJson bool `mapstructure:"no-json"`
	} `mapstructure:",squash"`

	ShutdownTimeout        time.Duration `mapstructure:"shutdown-timeout"`
	MetricsBindAddress     string        `mapstructure:"metric-bind-address"`
	HealthProbeBindAddress string        `mapstructure:"health-probe-bind-address"`

	LeaderElection struct {
		Enabled bool `mapstructure:"leader-elect"`
	} `mapstructure:",squash"`

	Sentry struct {
		Dsn string `mapstructure:"sentry-dsn"`
	} `mapstructure:",squash"`
}

func CommonFlags() *pflag.FlagSet {
	flagSet := pflag.NewFlagSet("common", pflag.ContinueOnError)

	flagSet.String("debug-label-value", "", "Set the debug label value")
	flagSet.Int("max-concurrent-reconciles", 1, "Set the max concurrent reconciles")
	flagSet.String("environment", "local", "Set the environment")
	flagSet.String("region", "local", "Set the region")
	flagSet.String("image-name", "", "Set the image name")
	flagSet.String("image-tag", "latest", "Set the image tag")
	flagSet.String("log-level", "info", "Set the log level")
	flagSet.Bool("no-json", false, "Disable JSON logging")
	flagSet.Duration("shutdown-timeout", 1*time.Minute, "Set the shutdown timeout")
	flagSet.String("metric-bind-address", ":8080", "Set the metrics bind address")
	flagSet.String("health-probe-bind-address", ":8090", "Set the health probe bind address")
	flagSet.Bool("leader-elect", false, "Enable leader election")
	flagSet.String("sentry-dsn", "", "Set the Sentry DSN")

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

func NewDefaultConfig(rootCmd *cobra.Command) (*viper.Viper, error) {
	v := viper.NewWithOptions(
		viper.EnvKeyReplacer(strings.NewReplacer("-", "_")),
	)

	v.AutomaticEnv()

	err := v.BindPFlags(CommonFlags())
	rootCmd.PersistentFlags().AddFlagSet(CommonFlags())

	return v, err
}

func BindConfigToFlags(v *viper.Viper, cmd *cobra.Command, config any) error {
	flagSet := generateFlagSet(config)
	err := v.BindPFlags(flagSet)
	if err != nil {
		return err
	}

	cmd.Flags().AddFlagSet(flagSet)

	return nil
}
