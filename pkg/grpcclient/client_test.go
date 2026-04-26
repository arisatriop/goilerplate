package grpcclient_test

import (
	"context"
	"net"
	"testing"
	"time"

	"goilerplate/pkg/grpcclient"
	pb "goilerplate/proto/hello/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type mockHelloServer struct {
	pb.UnimplementedHelloServiceServer
}

func (m *mockHelloServer) SayHello(_ context.Context, req *pb.SayHelloRequest) (*pb.SayHelloResponse, error) {
	return &pb.SayHelloResponse{Message: "Hello, " + req.Name}, nil
}

func startTestServer(t *testing.T) (addr string, stop func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)

	s := grpc.NewServer()
	pb.RegisterHelloServiceServer(s, &mockHelloServer{})

	go s.Serve(lis)

	return lis.Addr().String(), s.GracefulStop
}

func TestNewConn_UnaryCall(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	conn, err := grpcclient.NewConn(addr, grpcclient.Config{
		Timeout:  5 * time.Second,
		Insecure: true,
	})
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewHelloServiceClient(conn)
	resp, err := client.SayHello(context.Background(), &pb.SayHelloRequest{Name: "World"})
	assert.NoError(t, err)
	assert.Equal(t, "Hello, World", resp.Message)
}

func TestNewConn_DefaultTimeout(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	// Timeout = 0 should default to 30s without panicking
	conn, err := grpcclient.NewConn(addr, grpcclient.Config{Insecure: true})
	assert.NoError(t, err)
	defer conn.Close()

	client := pb.NewHelloServiceClient(conn)
	resp, err := client.SayHello(context.Background(), &pb.SayHelloRequest{Name: "Default"})
	assert.NoError(t, err)
	assert.Equal(t, "Hello, Default", resp.Message)
}

func TestNewConn_PropagatesRequestID(t *testing.T) {
	addr, stop := startTestServer(t)
	defer stop()

	conn, err := grpcclient.NewConn(addr, grpcclient.Config{
		Timeout:  5 * time.Second,
		Insecure: true,
	})
	assert.NoError(t, err)
	defer conn.Close()

	// Pre-set a request ID in outgoing metadata — interceptor should reuse it
	ctx := context.Background()
	client := pb.NewHelloServiceClient(conn)
	resp, err := client.SayHello(ctx, &pb.SayHelloRequest{Name: "ID"})
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Message)
}
