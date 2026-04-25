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
	foo := grpchandler.NewFoo()

	registry := grpcdelivery.NewServiceRegistry(
		hello,
		foo,
	)

	return &GrpcHandlers{
		ServiceRegistry: registry,
	}
}
