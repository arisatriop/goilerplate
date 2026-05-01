package grpcdelivery

import (
	grpchandler "goilerplate/internal/delivery/grpc/handler"

	barpb "github.com/arisatriop/goilerplate-proto/bar/v1"
	foopb "github.com/arisatriop/goilerplate-proto/foo/v1"
	hellopb "github.com/arisatriop/goilerplate-proto/hello/v1"

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
