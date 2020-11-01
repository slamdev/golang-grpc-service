package pkg

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	grpc_health_v1 "golang-grpc-service/api/grpc/health/v1"
	"google.golang.org/grpc"
	"net/http"
)

type ProxyServer interface {
	Start() error
	Stop() error
}

type proxyServer struct {
	srv    *http.Server
	client *grpc.ClientConn
}

type ProxyListener func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error

func NewProxyServer(httpPort int32, grpcPort int32, listeners ...ProxyListener) (ProxyServer, error) {
	client, err := grpc.Dial(fmt.Sprintf(":%d", grpcPort), grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to dialg grpc server; %w", err)
	}

	handler := runtime.NewServeMux()

	metricsHandler := func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	}

	if err := handler.HandlePath("GET", "/metrics", metricsHandler); err != nil {
		return nil, fmt.Errorf("failed register metrics handler; %w", err)
	}

	healthListener := func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
		return grpc_health_v1.RegisterHealthHandler(ctx, mux, conn)
	}

	listeners = append(listeners, healthListener)

	for _, listener := range listeners {
		if err := listener(context.TODO(), handler, client); err != nil {
			return nil, fmt.Errorf("failed register grpc listener; %w", err)
		}
	}

	return &proxyServer{
		srv: &http.Server{
			Addr:     fmt.Sprintf(":%d", httpPort),
			Handler:  handler,
			ErrorLog: zap.NewStdLog(zap.L()),
		},
		client: client,
	}, nil
}

func (s *proxyServer) Start() error {
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *proxyServer) Stop() error {
	srvErr := s.srv.Shutdown(context.TODO())
	clErr := s.client.Close()
	return multierr.Combine(srvErr, clErr)
}
