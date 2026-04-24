package bootstrap

import (
	"goilerplate/config"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func NewGrpcServer(cfg *config.Config) *grpc.Server {
	s := grpc.NewServer()

	if strings.ToLower(cfg.App.Env) != "production" {
		reflection.Register(s)
	}

	return s
}
