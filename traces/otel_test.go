package traces

import (
	"context"
	"net"
	"testing"
	"time"

	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
)

type mockExporter struct {
	sdkTrace.SpanExporter
}

func TestInitProvider_InvalidEndpoint(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Endpoint:       "invalid:0", // invalid endpoint to force error
	}
	_, err := InitProvider(ctx, cfg)
	if err == nil {
		t.Error("expected error for invalid endpoint, got nil")
	}
}

func TestConfig_initProvider_Success(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}
	exporter := &mockExporter{}
	shutdown, err := cfg.initProvider(ctx, exporter)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shutdown == nil {
		t.Error("expected shutdown function, got nil")
	}
}

func TestInitProvider_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()
	cfg := Config{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		Endpoint:       "localhost:65535", // unlikely to be open
	}
	_, err := InitProvider(ctx, cfg)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
}

func TestInitProvider_HappyPath(t *testing.T) {
	// Start a dummy gRPC server to simulate the OTLP collector
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	collectortrace.RegisterTraceServiceServer(s, &dummyTraceServer{})
	go s.Serve(lis)
	defer s.Stop()

	ctx := context.Background()
	cfg := Config{
		ServiceName:    "happy-service",
		ServiceVersion: "1.2.3",
		Endpoint:       lis.Addr().String(),
	}
	shutdown, err := InitProvider(ctx, cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shutdown == nil {
		t.Fatal("expected shutdown function, got nil")
	}
	// Call shutdown to ensure it works
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := shutdown(shutdownCtx); err != nil {
		t.Errorf("shutdown returned error: %v", err)
	}
}

// dummyTraceServer implements the OTLP TraceServiceServer interface with no-op methods.
type dummyTraceServer struct {
	collectortrace.UnimplementedTraceServiceServer
}
