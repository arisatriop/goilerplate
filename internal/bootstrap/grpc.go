package bootstrap

import (
	"goilerplate/config"
	grpcmiddleware "goilerplate/internal/delivery/grpc/middleware"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewGrpcServer creates the gRPC server. Call it only when grpc.enabled is true.
func NewGrpcServer(cfg *config.Config) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			grpcmiddleware.RequestLogger(),
			grpcmiddleware.Recovery(),
		),
	}
	if cfg.OTel.Enabled {
		opts = append(opts, grpc.StatsHandler(otelgrpc.NewServerHandler()))
	}

	s := grpc.NewServer(opts...)

	if strings.ToLower(cfg.App.Env) != "production" {
		reflection.Register(s)
	}

	return s
}
