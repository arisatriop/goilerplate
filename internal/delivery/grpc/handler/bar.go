package grpchandler

import (
	"context"

	bardomain "goilerplate/internal/domain/bar"
	"goilerplate/pkg/grpcresponse"
	pb "goilerplate/proto/bar/v1"
	"goilerplate/pkg/pagination"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Bar struct {
	pb.UnimplementedBarServiceServer
	uc bardomain.Usecase
}

func NewBar(uc bardomain.Usecase) *Bar {
	return &Bar{uc: uc}
}

func (b *Bar) CreateBar(ctx context.Context, req *pb.CreateBarRequest) (*pb.Bar, error) {
	entity := &bardomain.Bar{
		Code: req.Code,
		Bar:  req.Bar,
	}

	created, err := b.uc.Create(ctx, entity)
	if err != nil {
		return nil, grpcresponse.HandleError(ctx, err)
	}

	return toProtoBar(created), nil
}

func (b *Bar) GetBar(ctx context.Context, req *pb.GetBarRequest) (*pb.Bar, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	entity, err := b.uc.GetByID(ctx, req.Id)
	if err != nil {
		return nil, grpcresponse.HandleError(ctx, err)
	}

	return toProtoBar(entity), nil
}

func (b *Bar) ListBars(ctx context.Context, req *pb.ListBarsRequest) (*pb.ListBarsResponse, error) {
	filter := &bardomain.Filter{
		Keyword: req.Keyword,
		Pagination: &pagination.PaginationRequest{
			Page:  int(req.Page),
			Limit: int(req.Limit),
		},
	}
	filter.Pagination.Validate(pagination.DefaultPaginationConfig())

	bars, total, err := b.uc.GetList(ctx, filter)
	if err != nil {
		return nil, grpcresponse.HandleError(ctx, err)
	}

	items := make([]*pb.Bar, len(bars))
	for i, bar := range bars {
		items[i] = toProtoBar(bar)
	}

	return &pb.ListBarsResponse{
		Bars:  items,
		Total: total,
		Page:  int32(filter.Pagination.Page),
		Limit: int32(filter.Pagination.Limit),
	}, nil
}

func (b *Bar) UpdateBar(ctx context.Context, req *pb.UpdateBarRequest) (*pb.Bar, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	entity := &bardomain.Bar{
		ID:  req.Id,
		Code: req.Code,
		Bar:  req.Bar,
	}

	updated, err := b.uc.Update(ctx, entity)
	if err != nil {
		return nil, grpcresponse.HandleError(ctx, err)
	}

	return toProtoBar(updated), nil
}

func (b *Bar) DeleteBar(ctx context.Context, req *pb.DeleteBarRequest) (*emptypb.Empty, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	entity := &bardomain.Bar{ID: req.Id}

	if err := b.uc.Delete(ctx, entity); err != nil {
		return nil, grpcresponse.HandleError(ctx, err)
	}

	return &emptypb.Empty{}, nil
}

func toProtoBar(e *bardomain.Bar) *pb.Bar {
	return &pb.Bar{
		Id:   e.ID,
		Code: e.Code,
		Bar:  e.Bar,
	}
}
