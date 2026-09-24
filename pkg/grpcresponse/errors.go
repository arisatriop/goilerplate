package grpcresponse

import (
	"context"

	"goilerplate/pkg/apperr"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/logger"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorDomain names who issued an ErrorInfo, as google.rpc.ErrorInfo asks. Clients read the
// machine-readable code from ErrorInfo.Reason.
const ErrorDomain = "goilerplate"

// codeByKind is the one place a client error's category becomes a gRPC status code.
var codeByKind = map[apperr.Kind]codes.Code{
	apperr.Invalid:         codes.InvalidArgument,
	apperr.Unauthenticated: codes.Unauthenticated,
	apperr.Forbidden:       codes.PermissionDenied,
	apperr.NotFound:        codes.NotFound,
	apperr.Conflict:        codes.AlreadyExists,
}

// HandleError converts a use case error into a gRPC status, mirroring pkg/response.HandleError:
// a client error keeps its message and carries its code in an ErrorInfo detail (Google AIP-193);
// anything else is logged and answered with a generic Internal.
func HandleError(ctx context.Context, err error) error {
	if appErr, ok := apperr.As(err); ok {
		if code, known := codeByKind[appErr.Kind]; known {
			return withReason(status.New(code, appErr.Message), appErr.Code).Err()
		}
	}

	logger.Error(ctx, err)
	return status.Error(codes.Internal, constants.MsgInternalServerError)
}

// withReason attaches the machine-readable code. Attaching a detail only fails if it cannot be
// marshalled, which an ErrorInfo always can; the status without it is still a correct answer.
func withReason(st *status.Status, reason string) *status.Status {
	detailed, err := st.WithDetails(&errdetails.ErrorInfo{Reason: reason, Domain: ErrorDomain})
	if err != nil {
		return st
	}
	return detailed
}
