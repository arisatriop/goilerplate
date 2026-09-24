package grpcresponse_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"goilerplate/pkg/apperr"
	"goilerplate/pkg/grpcresponse"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestHandleError_ClientErrorCarriesItsCodeAsErrorInfo(t *testing.T) {
	tests := []struct {
		kind apperr.Kind
		want codes.Code
	}{
		{apperr.Invalid, codes.InvalidArgument},
		{apperr.Unauthenticated, codes.Unauthenticated},
		{apperr.Forbidden, codes.PermissionDenied},
		{apperr.NotFound, codes.NotFound},
		{apperr.Conflict, codes.AlreadyExists},
	}

	for _, tt := range tests {
		t.Run(tt.want.String(), func(t *testing.T) {
			err := fmt.Errorf("x: %w", apperr.New(tt.kind, "bar_problem", "Bar problem"))

			st, ok := status.FromError(grpcresponse.HandleError(context.Background(), err))

			require.True(t, ok)
			assert.Equal(t, tt.want, st.Code())
			assert.Equal(t, "Bar problem", st.Message())
			require.Len(t, st.Details(), 1)
			info, ok := st.Details()[0].(*errdetails.ErrorInfo)
			require.True(t, ok)
			assert.Equal(t, "bar_problem", info.GetReason())
			assert.Equal(t, grpcresponse.ErrorDomain, info.GetDomain())
		})
	}
}

func TestHandleError_ServerFaultIsAGenericInternal(t *testing.T) {
	st, _ := status.FromError(grpcresponse.HandleError(context.Background(), errors.New("pq: connection refused")))

	assert.Equal(t, codes.Internal, st.Code())
	assert.NotContains(t, st.Message(), "pq")
	assert.Empty(t, st.Details())
}
