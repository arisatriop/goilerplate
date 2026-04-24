package wire

import (
	grpcdelivery "goilerplate/internal/delivery/grpc"
	grpchandler "goilerplate/internal/delivery/grpc/handler"
)

type GrpcHandlers struct {
	ServiceRegistry *grpcdelivery.ServiceRegistry
}

func WireGrpcHandlers() *GrpcHandlers {
	hello := grpchandler.NewHello()

	registry := grpcdelivery.NewServiceRegistry(hello)

	return &GrpcHandlers{
		ServiceRegistry: registry,
	}
}
