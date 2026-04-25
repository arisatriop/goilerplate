package grpchandler

import (
	"context"
	pb "goilerplate/proto/bar/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Bar struct {
	pb.UnimplementedBarServiceServer
}

func NewBar() *Bar {
	return &Bar{}
}

func (b *Bar) CreateBar(_ context.Context, req *pb.CreateBarRequest) (*pb.CreateBarResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateBar not implemented")
}
