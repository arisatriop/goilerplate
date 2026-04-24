package grpchandler

import (
	"context"
	pb "goilerplate/proto/hello"
)

type Hello struct {
	pb.UnimplementedHelloServiceServer
}

func NewHello() *Hello {
	return &Hello{}
}

func (h *Hello) SayHello(_ context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{
		Message: "Hello, " + req.Name,
	}, nil
}
