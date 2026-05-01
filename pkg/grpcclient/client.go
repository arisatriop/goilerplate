package grpcclient

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"time"

	"goilerplate/pkg/constants"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const LogLabel = "outgoing-grpc-request-log"

type Config struct {
	Timeout  time.Duration
	Insecure bool
}

// NewConn creates a gRPC client connection with logging interceptors.
// Set Config.Insecure = true for local/dev; use grpc.WithTransportCredentials for prod TLS via extraOpts.
func NewConn(target string, cfg Config, extraOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	opts := []grpc.DialOption{
		grpc.WithUnaryInterceptor(loggingUnaryInterceptor(cfg.Timeout)),
		grpc.WithStreamInterceptor(loggingStreamInterceptor()),
	}

	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	opts = append(opts, extraOpts...)

	return grpc.NewClient(target, opts...)
}

func loggingUnaryInterceptor(defaultTimeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()

		requestID, ctx := ensureRequestID(ctx)

		if _, hasDeadline := ctx.Deadline(); !hasDeadline && defaultTimeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
			defer cancel()
		}

		err := invoker(ctx, method, req, reply, cc, opts...)

		duration := time.Since(start)
		st, _ := status.FromError(err)

		logAttrs := []slog.Attr{
			slog.String("label", LogLabel),
			slog.String("request_id", requestID),
			slog.String("method", method),
			slog.String("target", cc.Target()),
			slog.Any("request", marshalProto(req)),
			slog.Any("response", marshalProto(reply)),
			slog.String("status_code", st.Code().String()),
			slog.String("status_message", st.Message()),
			slog.Float64("latency_ms", float64(duration.Nanoseconds())/1e6),
		}

		if err != nil {
			logAttrs = append(logAttrs, slog.String("error", err.Error()))
		}

		if strings.ToLower(os.Getenv("APP_ENV")) != "local" {
			slog.LogAttrs(context.Background(), slog.LevelInfo, "Outgoing gRPC request", logAttrs...)
		}

		return err
	}
}

func loggingStreamInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		start := time.Now()

		requestID, ctx := ensureRequestID(ctx)

		stream, err := streamer(ctx, desc, cc, method, opts...)

		duration := time.Since(start)
		st, _ := status.FromError(err)

		logAttrs := []slog.Attr{
			slog.String("label", LogLabel),
			slog.String("request_id", requestID),
			slog.String("method", method),
			slog.String("target", cc.Target()),
			slog.String("stream_type", streamType(desc)),
			slog.String("status_code", st.Code().String()),
			slog.Float64("latency_ms", float64(duration.Nanoseconds())/1e6),
		}

		if err != nil {
			logAttrs = append(logAttrs, slog.String("error", err.Error()))
		}

		if strings.ToLower(os.Getenv("APP_ENV")) != "local" {
			slog.LogAttrs(context.Background(), slog.LevelInfo, "Outgoing gRPC stream", logAttrs...)
		}

		return stream, err
	}
}

func ensureRequestID(ctx context.Context) (string, context.Context) {
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		if ids := md.Get(constants.HeaderRequestID); len(ids) > 0 {
			return ids[0], ctx
		}
	}
	id := uuid.New().String()
	ctx = metadata.AppendToOutgoingContext(ctx, constants.HeaderRequestID, id)
	return id, ctx
}

func streamType(desc *grpc.StreamDesc) string {
	switch {
	case desc.ClientStreams && desc.ServerStreams:
		return "bidi"
	case desc.ClientStreams:
		return "client"
	case desc.ServerStreams:
		return "server"
	default:
		return "unary"
	}
}

func marshalProto(v any) any {
	if v == nil {
		return nil
	}
	msg, ok := v.(proto.Message)
	if !ok {
		return v
	}
	b, err := protojson.Marshal(msg)
	if err != nil {
		return nil
	}
	var m any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}
