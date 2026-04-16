package agent

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/user/practicum-metrics/internal/model"
	pb "github.com/user/practicum-metrics/internal/proto"
	"github.com/user/practicum-metrics/internal/retry"
	"github.com/user/practicum-metrics/internal/storage"
)

const metadataXRealIP = "x-real-ip"

// SetGRPCAddress enables gRPC batch sending to the provided server address.
func (ms *MetricsSender) SetGRPCAddress(address string) error {
	if address == "" {
		return nil
	}

	ms.realIP = resolveLocalIP(address)

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("create gRPC client: %w", err)
	}

	ms.grpcConn = conn
	ms.grpcClient = pb.NewMetricsClient(conn)
	return nil
}

func (ms *MetricsSender) Close() error {
	if ms.grpcConn == nil {
		return nil
	}
	return ms.grpcConn.Close()
}

func (ms *MetricsSender) sendMetricsBatchGRPC(metrics []model.Metrics) error {
	req := &pb.UpdateMetricsRequest{
		Metrics: make([]*pb.Metric, 0, len(metrics)),
	}

	for _, metric := range metrics {
		switch storage.MetricType(metric.MType) {
		case storage.Gauge:
			value := 0.0
			if metric.Value != nil {
				value = *metric.Value
			}
			req.Metrics = append(req.Metrics, &pb.Metric{
				Id:    metric.ID,
				Type:  pb.Metric_GAUGE,
				Value: value,
			})
		case storage.Counter:
			delta := int64(0)
			if metric.Delta != nil {
				delta = *metric.Delta
			}
			req.Metrics = append(req.Metrics, &pb.Metric{
				Id:    metric.ID,
				Type:  pb.Metric_COUNTER,
				Delta: delta,
			})
		default:
			return fmt.Errorf("unknown metric type: %s", metric.MType)
		}
	}

	return retry.Do(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
		defer cancel()

		if ms.realIP != "" {
			ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(metadataXRealIP, ms.realIP))
		}

		if _, err := ms.grpcClient.UpdateMetrics(ctx, req); err != nil {
			return fmt.Errorf("send metrics batch via gRPC: %w", err)
		}
		return nil
	})
}
