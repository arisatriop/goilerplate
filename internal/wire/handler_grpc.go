package wire

import (
	grpcdelivery "goilerplate/internal/delivery/grpc"
	grpchandler "goilerplate/internal/delivery/grpc/handler"
)

type GrpcHandlers struct {
	ServiceRegistry *grpcdelivery.ServiceRegistry
}

func WireGrpcHandlers(useCases *UseCases) *GrpcHandlers {
	hello := grpchandler.NewHello()
	foo := grpchandler.NewFoo()
	bar := grpchandler.NewBar(useCases.BarUC)

	registry := grpcdelivery.NewServiceRegistry(
		hello,
		foo,
		bar,
	)

	return &GrpcHandlers{
		ServiceRegistry: registry,
	}
}
