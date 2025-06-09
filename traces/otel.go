package traces

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	// ServiceName is the name of the instrumented library/service
	ServiceName string `mapstructure:"tracing-service-name"`
	// ServiceVersion is the version of the instrumented library/service
	// It must be in Semver format `<MAYOR>.<MINOR>.<PATCH>`
	ServiceVersion string `mapstructure:"tracing-service-version"`
	// Endpoint is the target of the collector.
	// Must be in the format `<DOMAIN>:<PORT>` without prefixed protocol
	// Ignored in the case of a LocalProvider
	Endpoint string `mapstructure:"tracing-endpoint"`
}

// InitProvider creates an OpenTelemetry provider for the concrete service.
// If the collector in the destination endpoint isn't reachable, then the init function will return an error.
func InitProvider(ctx context.Context, config Config) (func(ctx context.Context) error, error) {
	connCtx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	conn, err := grpc.DialContext(connCtx, config.Endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(connCtx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	return config.initProvider(ctx, traceExporter)
}

func (c Config) initProvider(ctx context.Context, exporter sdkTrace.SpanExporter) (func(ctx context.Context) error, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(c.ServiceName),
			semconv.ServiceVersion(c.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resouce: %w", err)
	}

	bsp := sdkTrace.NewBatchSpanProcessor(exporter)
	tracerProvider := sdkTrace.NewTracerProvider(
		sdkTrace.WithSampler(sdkTrace.AlwaysSample()),
		sdkTrace.WithResource(res),
		sdkTrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider.Shutdown, nil
}
