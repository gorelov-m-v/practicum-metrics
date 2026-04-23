package agent

import (
	"context"
	"net"
	"testing"

	"github.com/user/practicum-metrics/internal/grpcserver"
	"github.com/user/practicum-metrics/internal/model"
	"github.com/user/practicum-metrics/internal/service"
	"github.com/user/practicum-metrics/internal/storage"
)

func TestSendMetricsBatchGRPC(t *testing.T) {
	store := storage.NewMemStorage()
	metricsService := service.NewMetricsService(store, nil, nil, nil)
	_, subnet, err := net.ParseCIDR("127.0.0.0/8")
	if err != nil {
		t.Fatalf("parse subnet: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	server := grpcserver.NewServer(metricsService, nil, subnet)
	go func() {
		_ = server.Serve(listener)
	}()
	t.Cleanup(server.GracefulStop)

	sender := NewMetricsSender("http://127.0.0.1:8080", "")
	if err := sender.SetGRPCAddress(listener.Addr().String()); err != nil {
		t.Fatalf("set grpc address: %v", err)
	}
	t.Cleanup(func() {
		_ = sender.Close()
	})

	value := 7.5
	delta := int64(9)
	err = sender.SendMetricsBatch([]model.Metrics{
		{ID: "load", MType: string(storage.Gauge), Value: &value},
		{ID: "hits", MType: string(storage.Counter), Delta: &delta},
	})
	if err != nil {
		t.Fatalf("send metrics batch: %v", err)
	}

	gotValue, ok := store.GetGauge(context.Background(), "load")
	if !ok || gotValue != value {
		t.Fatalf("unexpected gauge: value=%v exists=%v", gotValue, ok)
	}

	gotDelta, ok := store.GetCounter(context.Background(), "hits")
	if !ok || gotDelta != delta {
		t.Fatalf("unexpected counter: value=%v exists=%v", gotDelta, ok)
	}
}
