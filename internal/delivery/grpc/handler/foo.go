package grpchandler

import (
	"context"

	pb "github.com/arisatriop/goilerplate-proto/foo/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Foo struct {
	pb.UnimplementedFooServiceServer
}

func NewFoo() *Foo {
	return &Foo{}
}

func (f *Foo) CreateFoo(_ context.Context, req *pb.CreateFooRequest) (*pb.CreateFooResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateFoo not implemented")
}

func (f *Foo) GetFoo(_ context.Context, req *pb.GetFooRequest) (*pb.GetFooResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetFoo not implemented")
}

func (f *Foo) ListFoos(_ context.Context, req *pb.ListFoosRequest) (*pb.ListFoosResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListFoos not implemented")
}

func (f *Foo) UpdateFoo(_ context.Context, req *pb.UpdateFooRequest) (*pb.UpdateFooResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateFoo not implemented")
}

func (f *Foo) DeleteFoo(_ context.Context, req *pb.DeleteFooRequest) (*pb.DeleteFooResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method DeleteFoo not implemented")
}
