package grpcresponse

import (
	"context"
	"errors"
	"net/http"

	"goilerplate/pkg/constants"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleError converts a domain error into a gRPC status error.
// Mirrors pkg/response.HandleError for the gRPC transport layer.
func HandleError(ctx context.Context, err error) error {
	var clientError *utils.ClientError
	if errors.As(err, &clientError) {
		return status.Error(httpCodeToGRPC(clientError.Code), clientError.Message)
	}

	logger.Error(ctx, err)
	return status.Error(codes.Internal, constants.MsgInternalServerError)
}

// httpCodeToGRPC maps HTTP status codes (used by ClientError) to gRPC codes.
func httpCodeToGRPC(httpCode int) codes.Code {
	switch httpCode {
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusUnprocessableEntity:
		return codes.InvalidArgument
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	case http.StatusServiceUnavailable:
		return codes.Unavailable
	default:
		return codes.Internal
	}
}
