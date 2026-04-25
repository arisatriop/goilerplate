package bootstrap

import (
	"goilerplate/config"
	grpcmiddleware "goilerplate/internal/delivery/grpc/middleware"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewGrpcServer(cfg *config.Config) *grpc.Server {
	s := grpc.NewServer(
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
