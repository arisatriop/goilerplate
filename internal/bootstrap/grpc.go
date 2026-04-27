package bootstrap

import (
	"goilerplate/config"
	grpcmiddleware "goilerplate/internal/delivery/grpc/middleware"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewGrpcServer(cfg *config.Config) *grpc.Server {
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			grpcmiddleware.RequestLogger(),
			grpcmiddleware.Recovery(),
		),
	)

	if strings.ToLower(cfg.App.Env) != "production" {
		reflection.Register(s)
	}

	return s
}
