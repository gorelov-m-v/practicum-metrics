package grpcserver

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/user/practicum-metrics/internal/proto"
	"github.com/user/practicum-metrics/internal/service"
	"github.com/user/practicum-metrics/internal/storage"
)

func TestMetricsServer_UpdateMetrics(t *testing.T) {
	store := storage.NewMemStorage()
	metricsService := service.NewMetricsService(store, nil, nil, nil)
	server := NewMetricsServer(metricsService, nil)

	req := pb.UpdateMetricsRequest_builder{
		Metrics: []*pb.Metric{
			pb.Metric_builder{Id: "temperature", Type: pb.Metric_GAUGE, Value: 42.5}.Build(),
			pb.Metric_builder{Id: "requests", Type: pb.Metric_COUNTER, Delta: 3}.Build(),
		},
	}.Build()

	_, err := server.UpdateMetrics(context.Background(), req)
	if err != nil {
		t.Fatalf("update metrics: %v", err)
	}

	gauge, ok := store.GetGauge(context.Background(), "temperature")
	if !ok || gauge != 42.5 {
		t.Fatalf("unexpected gauge: value=%v exists=%v", gauge, ok)
	}

	counter, ok := store.GetCounter(context.Background(), "requests")
	if !ok || counter != 3 {
		t.Fatalf("unexpected counter: value=%v exists=%v", counter, ok)
	}
}

func TestTrustedSubnetInterceptor(t *testing.T) {
	_, subnet, err := net.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatalf("parse subnet: %v", err)
	}

	interceptor := TrustedSubnetInterceptor(subnet)
	handler := func(context.Context, interface{}) (interface{}, error) {
		return "ok", nil
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(metadataXRealIP, "192.168.1.42"))
	resp, err := interceptor(ctx, nil, nil, handler)
	if err != nil {
		t.Fatalf("expected allowed request, got error: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("unexpected response: %v", resp)
	}

	ctx = metadata.NewIncomingContext(context.Background(), metadata.Pairs(metadataXRealIP, "10.0.0.1"))
	_, err = interceptor(ctx, nil, nil, handler)
	if status.Code(err) != codes.PermissionDenied {
		t.Fatalf("expected PermissionDenied, got %v", err)
	}
}
