package grpcmiddleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"time"

	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const LogLabel = "incoming-grpc-request-log"

func RequestLogger() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := utils.Now()
		startTime := start.Format(time.RFC3339Nano)

		requestID := extractOrGenerateRequestID(ctx)
		ctx = context.WithValue(ctx, constants.ContextKeyRequestID, requestID)

		peerAddr := ""
		if p, ok := peer.FromContext(ctx); ok {
			peerAddr = p.Addr.String()
		}

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		endTime := utils.Now().Format(time.RFC3339Nano)
		st, _ := status.FromError(err)

		logAttrs := []slog.Attr{
			slog.String("label", LogLabel),
			slog.String("request_id", requestID),
			slog.String("method", info.FullMethod),
			slog.String("peer_address", peerAddr),
			slog.Any("request", marshalProto(req)),
			slog.Any("response", marshalProto(resp)),
			slog.String("status_code", st.Code().String()),
			slog.String("status_message", st.Message()),
			slog.String("start_time", startTime),
			slog.String("end_time", endTime),
			slog.Float64("latency_ms", float64(duration.Nanoseconds())/1e6),
		}

		if err != nil {
			logAttrs = append(logAttrs, slog.String("error", err.Error()))
		}

		if strings.ToLower(os.Getenv("APP_ENV")) != "local" {
			slog.LogAttrs(ctx, slog.LevelInfo, "Incoming gRPC request", logAttrs...)
		}

		return resp, err
	}
}

func extractOrGenerateRequestID(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ids := md.Get(strings.ToLower(constants.HeaderRequestID)); len(ids) > 0 {
			return ids[0]
		}
	}
	return uuid.New().String()
}

// marshalProto serializes a proto message to a loggable map (duplicated from grpcclient for independence)
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
	json.Unmarshal(b, &m)
	return m
}
