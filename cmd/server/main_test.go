package main

import (
	"context"
	"io"
	"log/slog"
	"net"
	"testing"
	"time"

	"goilerplate/internal/bootstrap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func quietApp(server *grpc.Server) *bootstrap.App {
	return &bootstrap.App{
		Log:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		GrpcServer: server,
	}
}

// serveBlocking starts a gRPC server whose every method hangs until the returned release
// function is called, so a test can hold one RPC open across a drain.
func serveBlocking(t *testing.T) (server *grpc.Server, addr string, release func()) {
	t.Helper()

	held := make(chan struct{})
	var releaseOnce bool
	release = func() {
		if !releaseOnce {
			releaseOnce = true
			close(held)
		}
	}

	server = grpc.NewServer(grpc.UnknownServiceHandler(
		func(any, grpc.ServerStream) error {
			<-held
			return nil
		}))

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(release)

	return server, listener.Addr().String(), release
}

// GracefulStop waits for every active RPC with no deadline of its own. One long-lived stream
// would otherwise keep the process alive past terminationGracePeriodSeconds and get it
// SIGKILLed instead of letting it exit.
func TestDrainGRPC_BoundedByTheContext(t *testing.T) {
	server, addr, release := serveBlocking(t)
	defer release()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	// Fire a call that the handler will never finish, and wait until the server has it.
	callStarted := make(chan struct{})
	go func() {
		close(callStarted)
		_ = conn.Invoke(context.Background(), "/test.Service/Hang", &emptyMessage{}, &emptyMessage{})
	}()
	<-callStarted
	time.Sleep(200 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	start := time.Now()
	go func() {
		drainGRPC(ctx, quietApp(server))
		close(done)
	}()

	select {
	case <-done:
		assert.Less(t, time.Since(start), 5*time.Second,
			"drain should end at the deadline, not wait for the RPC")
	case <-time.After(5*time.Second + time.Second):
		t.Fatal("drainGRPC did not return: GracefulStop was not bounded by the context")
	}
}

// With nothing in flight the drain must finish on its own, without the forced Stop path.
func TestDrainGRPC_ReturnsImmediatelyWhenIdle(t *testing.T) {
	server, _, release := serveBlocking(t)
	defer release()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	start := time.Now()
	drainGRPC(ctx, quietApp(server))

	assert.Less(t, time.Since(start), 2*time.Second)
}

// emptyMessage is a minimal proto payload: the unknown-service handler never decodes it, so it
// only has to satisfy the codec's interface.
type emptyMessage struct{}

func (*emptyMessage) Reset()         {}
func (*emptyMessage) String() string { return "" }
func (*emptyMessage) ProtoMessage()  {}
