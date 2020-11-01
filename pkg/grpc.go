package pkg

import (
	"context"
	"fmt"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"golang-grpc-service/api/grpc/health/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
)

type GrpcServer interface {
	Start() error
	Stop() error
}

type grpcServer struct {
	srv  *grpc.Server
	port int32
}

type GrpcListener func(srv *grpc.Server)

func NewGrpcServer(port int32, listeners ...GrpcListener) GrpcServer {
	grpc_prometheus.EnableHandlingTimeHistogram()

	skipHealthLogging := grpc_zap.WithDecider(func(fullMethodName string, err error) bool {
		return fullMethodName != "/grpc.health.v1.Health/Check"
	})

	srv := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_prometheus.StreamServerInterceptor,
			grpc_recovery.StreamServerInterceptor(),
			grpc_zap.StreamServerInterceptor(zap.L()),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_recovery.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(zap.L(), skipHealthLogging),
		)),
	)

	for _, listener := range listeners {
		listener(srv)
	}

	grpc_health_v1.RegisterHealthServer(srv, healthService{})

	grpc_prometheus.Register(srv)

	return &grpcServer{
		srv:  srv,
		port: port,
	}
}

func (s *grpcServer) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen %d; %w", s.port, err)
	}
	return s.srv.Serve(lis)
}

func (s *grpcServer) Stop() error {
	s.srv.GracefulStop()
	return nil
}

type healthService struct {
	grpc_health_v1.UnimplementedHealthServer
}

func (healthService) Check(context.Context, *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}

func (healthService) Watch(*grpc_health_v1.HealthCheckRequest, grpc_health_v1.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watching is not supported")
}
