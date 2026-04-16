package grpcserver

import (
	"context"
	"fmt"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/user/practicum-metrics/internal/model"
	pb "github.com/user/practicum-metrics/internal/proto"
	"github.com/user/practicum-metrics/internal/service"
	"github.com/user/practicum-metrics/internal/storage"
)

const metadataXRealIP = "x-real-ip"

type MetricsServer struct {
	pb.UnimplementedMetricsServer

	service   *service.MetricsService
	persister *storage.Persister
}

func NewServer(service *service.MetricsService, persister *storage.Persister, trustedSubnet *net.IPNet) *grpc.Server {
	server := grpc.NewServer(grpc.UnaryInterceptor(TrustedSubnetInterceptor(trustedSubnet)))
	pb.RegisterMetricsServer(server, NewMetricsServer(service, persister))
	return server
}

func NewMetricsServer(service *service.MetricsService, persister *storage.Persister) *MetricsServer {
	return &MetricsServer{
		service:   service,
		persister: persister,
	}
}

func (s *MetricsServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	if len(req.GetMetrics()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty batch not allowed")
	}

	metrics := make([]model.Metrics, 0, len(req.GetMetrics()))
	for _, metric := range req.GetMetrics() {
		switch metric.GetType() {
		case pb.Metric_GAUGE:
			value := metric.GetValue()
			metrics = append(metrics, model.Metrics{
				ID:    metric.GetId(),
				MType: string(storage.Gauge),
				Value: &value,
			})
		case pb.Metric_COUNTER:
			delta := metric.GetDelta()
			metrics = append(metrics, model.Metrics{
				ID:    metric.GetId(),
				MType: string(storage.Counter),
				Delta: &delta,
			})
		default:
			return nil, status.Errorf(codes.InvalidArgument, "unknown metric type: %s", metric.GetType().String())
		}
	}

	if err := s.service.UpdateMetricsBatch(ctx, metrics); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if s.persister != nil && s.persister.IsSyncMode() {
		s.persister.SaveSync()
	}

	return &pb.UpdateMetricsResponse{}, nil
}

func TrustedSubnetInterceptor(trustedSubnet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if trustedSubnet == nil {
			return handler(ctx, req)
		}

		ip, err := realIPFromMetadata(ctx)
		if err != nil {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}

		if !trustedSubnet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "IP address is outside trusted subnet")
		}

		return handler(ctx, req)
	}
}

func realIPFromMetadata(ctx context.Context) (net.IP, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing %s metadata", metadataXRealIP)
	}

	values := md.Get(metadataXRealIP)
	if len(values) == 0 {
		return nil, fmt.Errorf("missing %s metadata", metadataXRealIP)
	}

	ip := net.ParseIP(strings.TrimSpace(values[0]))
	if ip == nil {
		return nil, fmt.Errorf("invalid %s metadata", metadataXRealIP)
	}

	return ip, nil
}
