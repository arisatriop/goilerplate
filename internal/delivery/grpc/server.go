package grpcdelivery

import (
	grpchandler "goilerplate/internal/delivery/grpc/handler"
	barpb "goilerplate/proto/bar/v1"
	foopb "goilerplate/proto/foo/v1"
	hellopb "goilerplate/proto/hello/v1"

	"google.golang.org/grpc"
)

type ServiceRegistry struct {
	Hello *grpchandler.Hello
	Foo   *grpchandler.Foo
	Bar   *grpchandler.Bar
}

func NewServiceRegistry(
	hello *grpchandler.Hello,
	foo *grpchandler.Foo,
	bar *grpchandler.Bar,
) *ServiceRegistry {
	return &ServiceRegistry{
		Hello: hello,
		Foo:   foo,
		Bar:   bar,
	}
}

func (r *ServiceRegistry) Register(s *grpc.Server) {
	hellopb.RegisterHelloServiceServer(s, r.Hello)
	foopb.RegisterFooServiceServer(s, r.Foo)
	barpb.RegisterBarServiceServer(s, r.Bar)
}
