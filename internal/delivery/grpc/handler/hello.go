package grpchandler

import (
	"context"
	pb "goilerplate/proto/hello/v1"
)

type Hello struct {
	pb.UnimplementedHelloServiceServer
}

func NewHello() *Hello {
	return &Hello{}
}

func (h *Hello) SayHello(_ context.Context, req *pb.SayHelloRequest) (*pb.SayHelloResponse, error) {
	return &pb.SayHelloResponse{
		Message: "Hello, " + req.Name,
	}, nil
}
