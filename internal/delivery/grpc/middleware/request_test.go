package grpcmiddleware

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"goilerplate/pkg/redact"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestRequestLogger_RedactsPayloads(t *testing.T) {
	// Arrange
	t.Setenv("APP_ENV", "test")
	logs := &bytes.Buffer{}
	original := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(logs, nil)))
	t.Cleanup(func() { slog.SetDefault(original) })

	req, err := structpb.NewStruct(map[string]any{"email": "user@example.com", "password": "hunter2-password"})
	require.NoError(t, err)
	resp, err := structpb.NewStruct(map[string]any{"accessToken": "eyJaccess.secret.token"})
	require.NoError(t, err)

	handler := func(context.Context, any) (any, error) { return resp, nil }
	info := &grpc.UnaryServerInfo{FullMethod: "/auth.v1.AuthService/Login"}

	// Act
	_, err = RequestLogger()(context.Background(), req, info, handler)

	// Assert
	require.NoError(t, err)
	output := logs.String()
	assert.Contains(t, output, LogLabel)
	assert.NotContains(t, output, "hunter2-password")
	assert.NotContains(t, output, "eyJaccess.secret.token")
	assert.Contains(t, output, redact.Mask)
	assert.Contains(t, output, "user@example.com")
}
