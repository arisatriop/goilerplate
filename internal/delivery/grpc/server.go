package grpcdelivery

import (
	grpchandler "goilerplate/internal/delivery/grpc/handler"
	pb "goilerplate/proto/hello/v1"

	"google.golang.org/grpc"
)

type ServiceRegistry struct {
	Hello *grpchandler.Hello
}

func NewServiceRegistry(hello *grpchandler.Hello) *ServiceRegistry {
	return &ServiceRegistry{
		Hello: hello,
	}
}

func (r *ServiceRegistry) Register(s *grpc.Server) {
	pb.RegisterHelloServiceServer(s, r.Hello)
}
