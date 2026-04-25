package grpcmiddleware

import (
	"context"
	"log/slog"
	"runtime/debug"

	"goilerplate/pkg/constants"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Recovery() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				requestID := ""
				if id := ctx.Value(constants.ContextKeyRequestID); id != nil {
					requestID = id.(string)
				}

				slog.ErrorContext(ctx,
					"Panic recovered in gRPC handler",
					slog.String("request_id", requestID),
					slog.String("method", info.FullMethod),
					slog.Any("panic_value", r),
					slog.String("stack_trace", string(debug.Stack())),
				)

				err = status.Error(codes.Internal, "An unexpected error occurred")
			}
		}()

		return handler(ctx, req)
	}
}
