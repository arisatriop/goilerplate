package bootstrap

import (
	"fmt"

	"goilerplate/config"
	grpcmiddleware "goilerplate/internal/delivery/grpc/middleware"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

// NewGrpcServer creates the gRPC server. Call it only when grpc.enabled is true.
//
// It is built during wiring rather than at bootstrap, because interceptors can only be given to
// grpc.NewServer and the auth interceptor needs the token validator and session store — which
// do not exist until the infrastructure layer is wired.
func NewGrpcServer(cfg *config.Config, authInterceptor *grpcmiddleware.Auth) (*grpc.Server, error) {
	opts := []grpc.ServerOption{
		// Auth runs after the request logger, so a rejected call is still logged, and before
		// recovery is irrelevant — a call that fails auth never reaches a handler to panic.
		grpc.ChainUnaryInterceptor(
			grpcmiddleware.RequestLogger(),
			grpcmiddleware.Recovery(),
			authInterceptor.Unary(),
		),
		grpc.ChainStreamInterceptor(
			authInterceptor.Stream(),
		),
	}

	if cfg.GRPC.TLS.Enabled {
		creds, err := credentials.NewServerTLSFromFile(cfg.GRPC.TLS.CertFile, cfg.GRPC.TLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("loading grpc TLS certificate: %w", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	if cfg.OTel.Enabled {
		opts = append(opts, grpc.StatsHandler(otelgrpc.NewServerHandler()))
	}

	s := grpc.NewServer(opts...)

	// Off unless asked for, rather than on outside production: reflection publishes the whole
	// service surface to anyone who can reach the port, and "not production" is not the same
	// question as "safe to enumerate".
	if cfg.GRPC.Reflection {
		reflection.Register(s)
	}

	return s, nil
}
