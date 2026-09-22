package main

import (
	"context"
	"io"
	"log/slog"
	"net"
	"sync"
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
// function is called, so a test can hold one RPC open across a drain. entered is closed once a
// call has actually reached the handler, which is what a test must wait for — sleeping instead
// lets the drain start before the RPC lands, and then it is GracefulStop finishing on its own
// that makes the test pass, not the code under test.
func serveBlocking(t *testing.T) (server *grpc.Server, addr string, entered <-chan struct{}, release func()) {
	t.Helper()

	held := make(chan struct{})
	release = sync.OnceFunc(func() { close(held) })
	t.Cleanup(release)

	arrived := make(chan struct{})
	announce := sync.OnceFunc(func() { close(arrived) })

	server = grpc.NewServer(grpc.UnknownServiceHandler(
		func(any, grpc.ServerStream) error {
			announce()
			<-held
			return nil
		}))

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() { _ = server.Serve(listener) }()

	return server, listener.Addr().String(), arrived, release
}

// GracefulStop waits for every active RPC with no deadline of its own. One long-lived stream
// would otherwise keep the process alive past terminationGracePeriodSeconds and get it
// SIGKILLed instead of letting it exit.
func TestDrainGRPC_BoundedByTheContext(t *testing.T) {
	server, addr, entered, release := serveBlocking(t)
	defer release()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	// Fire a call the handler will never finish, and wait until it has genuinely arrived.
	go func() {
		_ = conn.Invoke(context.Background(), "/test.Service/Hang", &emptyMessage{}, &emptyMessage{})
	}()
	select {
	case <-entered:
	case <-time.After(10 * time.Second):
		t.Fatal("the RPC never reached the handler; the test would not be exercising anything")
	}

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
	server, _, _, release := serveBlocking(t)
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
