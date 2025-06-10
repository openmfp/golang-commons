package config

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/openmfp/golang-commons/context/keys"
	"github.com/openmfp/golang-commons/traces"
)

func SetConfigInContext(ctx context.Context, config any) context.Context {
	return context.WithValue(ctx, keys.ConfigCtxKey, config)
}

func LoadConfigFromContext(ctx context.Context) any {
	return ctx.Value(keys.ConfigCtxKey)
}

type CommonServiceConfig struct {
	DebugLabelValue         string `mapstructure:"debug-label-value" default:"" description:"Set the debug label value"`
	MaxConcurrentReconciles int    `mapstructure:"max-concurrent-reconciles" default:"10" description:"Set the max concurrent reconciles"`
	Environment             string `mapstructure:"environment"`
	Region                  string `mapstructure:"region" default:"local" description:"Set the region of the service, e.g. local, dev, staging, prod"`
	Kubeconfig              string `mapstructure:"kubeconfig" description:"Set the kubeconfig path"`
	IsLocal                 bool   `mapstructure:"is-local" default:"false" description:"Flagging execution to be local"`

	Image struct {
		Name string `mapstructure:"image-name" description:"Set the image name"`
		Tag  string `mapstructure:"image-tag" description:"Set the image tag"`
	} `mapstructure:",squash"`

	Log struct {
		Level  string `mapstructure:"log-level" default:"info" description:"Set the log level"`
		NoJson bool   `mapstructure:"no-json" default:"false" description:"Disable JSON logging"`
	} `mapstructure:",squash"`

	ShutdownTimeout time.Duration `mapstructure:"shutdown-timeout" default:"1m" description:"Set the shutdown timeout as duration in seconds, e.g. 30s, 1m, 2h"`
	Metrics         struct {
		BindAddress string `mapstructure:"metrics-bind-address" default:":9090" description:"Set the metrics bind address"`
		Secure      bool   `mapstructure:"metrics-secure" default:"false" description:"Set if metrics should be exposed via https"`
	} `mapstructure:",squash"`
	Tracing struct {
		Enabled   bool          `mapstructure:"tracing-enabled" default:"false" description:"Enable tracing for the service"`
		Collector traces.Config `mapstructure:",squash"`
	} `mapstructure:",squash"`
	EnableHTTP2            bool   `mapstructure:"enable-http2" default:"true" description:"Toggle to disable metrics/webhook serving using http2"`
	HealthProbeBindAddress string `mapstructure:"health-probe-bind-address" default:":8090" description:"Set the health probe bind address"`

	LeaderElection struct {
		Enabled bool `mapstructure:"leader-elect" default:"false" description:"Enable leader election for the controller manager"`
	} `mapstructure:",squash"`

	Sentry struct {
		Dsn string `mapstructure:"sentry-dsn" description:"Set the Sentry DSN for error reporting"`
	} `mapstructure:",squash"`
}

// generateFlagSet generates a pflag.FlagSet from a struct based on its `mapstructure` tags.
func generateFlagSet(config any) (*pflag.FlagSet, error) {
	flagSet := pflag.NewFlagSet("generated", pflag.ContinueOnError)
	err := traverseStruct(reflect.ValueOf(config), flagSet, "")
	return flagSet, err
}

// traverseStruct recursively traverses a struct and adds flags to the FlagSet.
func traverseStruct(value reflect.Value, flagSet *pflag.FlagSet, prefix string) error {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return errors.New("value must be a struct")
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

		defaultValueTag := field.Tag.Get("default")
		defaultStrValue := ""
		if defaultValueTag != "" {
			defaultStrValue = defaultValueTag
		}

		descriptionValueTag := field.Tag.Get("description")
		descriptionStrValue := ""
		if descriptionValueTag != "" {
			descriptionStrValue = descriptionValueTag
		}

		// Handle nested structs
		if fieldValue.Kind() == reflect.Struct {
			if tag == ",squash" {
				err := traverseStruct(fieldValue, flagSet, "")
				if err != nil {
					return err
				}
			} else {
				err := traverseStruct(fieldValue, flagSet, prefix+tag+".")
				if err != nil {
					return err
				}
			}
			continue
		}

		description := fmt.Sprintf("Set the %s", tag)
		if descriptionStrValue != "" {
			description = descriptionStrValue
		}

		// Add flags based on the field type
		switch fieldValue.Kind() {
		case reflect.String:
			flagSet.String(prefix+tag, defaultStrValue, description)
		case reflect.Int, reflect.Int64:
			i := 0
			if defaultStrValue != "" {
				parsedInt, err := strconv.Atoi(defaultStrValue)
				if err != nil {
					return err
				}
				i = parsedInt
			}
			if fieldValue.Type() == reflect.TypeOf(time.Duration(0)) {
				durVal := time.Duration(i) * time.Second
				durDescription := fmt.Sprintf("Set the %s in seconds", tag)
				if descriptionStrValue != "" {
					durDescription = descriptionStrValue
				}
				flagSet.Duration(prefix+tag, durVal, durDescription)
			} else {
				flagSet.Int(prefix+tag, i, description)
			}
		case reflect.Bool:
			b, err := strconv.ParseBool(defaultStrValue)
			if err != nil {
				return err
			}
			flagSet.Bool(prefix+tag, b, description)
		default:
			return fmt.Errorf("unsupported field type %s for field %s", fieldValue.Kind(), field.Name)
		}

	}
	return nil
}

func NewDefaultConfig(rootCmd *cobra.Command) (*viper.Viper, *CommonServiceConfig, error) {
	v := viper.NewWithOptions(
		viper.EnvKeyReplacer(strings.NewReplacer("-", "_")),
	)

	v.AutomaticEnv()

	var config CommonServiceConfig
	flagSet, err := generateFlagSet(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate flag set: %w", err)
	}

	err = v.BindPFlags(flagSet)
	rootCmd.PersistentFlags().AddFlagSet(flagSet)

	cobra.OnInitialize(unmarshalIntoStruct(v, &config))

	return v, &config, err
}

func BindConfigToFlags(v *viper.Viper, cmd *cobra.Command, config any) error {
	flagSet, err := generateFlagSet(config)
	if err != nil {
		return fmt.Errorf("failed to generate flag set: %w", err)
	}
	err = v.BindPFlags(flagSet)
	if err != nil {
		return err
	}

	cmd.Flags().AddFlagSet(flagSet)

	cobra.OnInitialize(unmarshalIntoStruct(v, config))

	return nil
}

func unmarshalIntoStruct(v *viper.Viper, cfg any) func() {
	return func() {
		if err := v.Unmarshal(cfg); err != nil {
			panic(fmt.Errorf("failed to unmarshal config: %w", err))
		}
	}
}
