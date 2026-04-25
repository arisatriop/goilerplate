package grpcdelivery

import (
	grpchandler "goilerplate/internal/delivery/grpc/handler"
	foopb "goilerplate/proto/foo/v1"
	hellopb "goilerplate/proto/hello/v1"

	"google.golang.org/grpc"
)

type ServiceRegistry struct {
	Hello *grpchandler.Hello
	Foo   *grpchandler.Foo
}

func NewServiceRegistry(
	hello *grpchandler.Hello,
	foo *grpchandler.Foo,
) *ServiceRegistry {
	return &ServiceRegistry{
		Hello: hello,
		Foo:   foo,
	}
}

func (r *ServiceRegistry) Register(s *grpc.Server) {
	hellopb.RegisterHelloServiceServer(s, r.Hello)
	foopb.RegisterFooServiceServer(s, r.Foo)
}
