package grpchandler

import (
	"context"

	"goilerplate/internal/domain/bar"
	"goilerplate/pkg/grpcresponse"
	pb "goilerplate/proto/bar/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Bar struct {
	pb.UnimplementedBarServiceServer
	uc bar.Usecase
}

func NewBar(uc bar.Usecase) *Bar {
	return &Bar{uc: uc}
}

func (b *Bar) CreateBar(ctx context.Context, req *pb.CreateBarRequest) (*pb.CreateBarResponse, error) {
	entity := &bar.Bar{
		Code: req.Code,
		Bar:  req.Bar,
	}

	if err := b.uc.Create(ctx, entity); err != nil {
		return nil, grpcresponse.HandleError(ctx, err)
	}

	return &pb.CreateBarResponse{}, nil
}

func (b *Bar) GetBar(_ context.Context, req *pb.GetBarRequest) (*pb.GetBarResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method GetBar not implemented")
}

func (b *Bar) ListBars(_ context.Context, req *pb.ListBarsRequest) (*pb.ListBarsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method ListBars not implemented")
}

func (b *Bar) UpdateBar(_ context.Context, req *pb.UpdateBarRequest) (*pb.UpdateBarResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method UpdateBar not implemented")
}

func (b *Bar) DeleteBar(_ context.Context, req *pb.DeleteBarRequest) (*pb.DeleteBarResponse, error) {
	return nil, status.Error(codes.Unimplemented, "method DeleteBar not implemented")
}
