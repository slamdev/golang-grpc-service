package internal

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"golang-grpc-service/api"
	"golang-grpc-service/pkg"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type App interface {
	Start() error
	Stop() error
}

type app struct {
	config      Config
	grpcServer  pkg.GrpcServer
	proxyServer pkg.ProxyServer
}

func NewApp() (App, error) {
	app := app{}
	if err := pkg.PopulateConfig(&app.config); err != nil {
		return nil, fmt.Errorf("failed to populate config; %w", err)
	}

	if err := pkg.ConfigureLogger(app.config.Logger.Production); err != nil {
		return nil, fmt.Errorf("failed to configure logger; %w", err)
	}

	zap.L().Info("starting app", zap.Reflect("config", app.config))

	app.grpcServer = pkg.NewGrpcServer(app.config.Grpc.Port, func(s *grpc.Server) {
		api.RegisterGolangGrpcServiceServer(s, NewController())
	})

	var err error
	app.proxyServer, err = pkg.NewProxyServer(app.config.Http.Port, app.config.Grpc.Port, func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
		return api.RegisterGolangGrpcServiceHandler(ctx, mux, conn)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy server; %w", err)
	}

	return &app, nil
}

func (a *app) Start() error {
	wg, _ := errgroup.WithContext(context.TODO())
	wg.Go(func() error { return a.grpcServer.Start() })
	wg.Go(func() error { return a.proxyServer.Start() })
	return wg.Wait()
}

func (a *app) Stop() error {
	return multierr.Combine(a.proxyServer.Stop(), a.grpcServer.Stop())
}

type Config struct {
	Grpc struct {
		Port int32
	}
	Http struct {
		Port int32
	}
	Logger struct {
		Production bool
	}
}
