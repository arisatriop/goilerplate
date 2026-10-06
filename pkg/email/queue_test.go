package email_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"goilerplate/pkg/email"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingSender records what it was asked to send. When gate is set, each Send waits for it
// or for its context, standing in for a slow provider.
type recordingSender struct {
	mu      sync.Mutex
	sent    []email.Message
	started chan struct{}
	gate    chan struct{}
	err     error
}

func (s *recordingSender) Send(ctx context.Context, msg email.Message) error {
	if s.started != nil {
		s.started <- struct{}{}
	}
	if s.gate != nil {
		select {
		case <-s.gate:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent = append(s.sent, msg)
	return s.err
}

func (s *recordingSender) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.sent)
}

func message(subject string) email.Message {
	return email.Message{To: "user@example.org", Subject: subject, Text: "body"}
}

func TestQueue_ShutdownDeliversWhatWasQueued(t *testing.T) {
	sender := &recordingSender{}
	queue := email.NewQueue(sender, email.QueueOptions{})

	for _, subject := range []string{"one", "two", "three"} {
		require.NoError(t, queue.Send(t.Context(), message(subject)))
	}
	require.NoError(t, queue.Shutdown(t.Context()))

	assert.Equal(t, 3, sender.count(), "a graceful shutdown drains the queue")
	assert.ErrorIs(t, queue.Send(t.Context(), message("late")), email.ErrQueueClosed)
	assert.NoError(t, queue.Shutdown(t.Context()), "shutting down twice is harmless")
}

// A stalled provider must not back up into request handling: Send refuses instead of blocking.
func TestQueue_FullQueueRefusesInsteadOfBlocking(t *testing.T) {
	sender := &recordingSender{started: make(chan struct{}, 1), gate: make(chan struct{})}
	queue := email.NewQueue(sender, email.QueueOptions{Size: 1})
	t.Cleanup(func() {
		close(sender.gate)
		_ = queue.Shutdown(context.Background())
	})

	require.NoError(t, queue.Send(t.Context(), message("in flight")))
	<-sender.started // the worker holds it; the buffer is empty again
	require.NoError(t, queue.Send(t.Context(), message("buffered")))

	err := queue.Send(t.Context(), message("overflow"))

	assert.ErrorIs(t, err, email.ErrQueueFull)
}

func TestQueue_ShutdownGivesUpWhenItsBudgetEnds(t *testing.T) {
	sender := &recordingSender{started: make(chan struct{}, 1), gate: make(chan struct{})}
	queue := email.NewQueue(sender, email.QueueOptions{})

	require.NoError(t, queue.Send(t.Context(), message("stuck")))
	require.NoError(t, queue.Send(t.Context(), message("waiting")))
	<-sender.started

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	err := queue.Shutdown(ctx)

	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Contains(t, err.Error(), "2 message(s) not delivered")
	assert.Zero(t, sender.count())
}

// Nobody waits on a queued delivery, so its failure has to reach the log or it is lost.
func TestQueue_LogsAFailedDeliveryWithoutTheRecipient(t *testing.T) {
	var logs bytes.Buffer
	sender := &recordingSender{err: errors.New("provider said no")}
	queue := email.NewQueue(sender, email.QueueOptions{Logger: slog.New(slog.NewTextHandler(&logs, nil))})

	require.NoError(t, queue.Send(t.Context(), message("Reset your password")))
	require.NoError(t, queue.Shutdown(t.Context()))

	assert.Contains(t, logs.String(), "provider said no")
	assert.Contains(t, logs.String(), "Reset your password")
	assert.NotContains(t, logs.String(), "user@example.org")
}

func TestQueue_RefusesAMalformedMessageImmediately(t *testing.T) {
	queue := email.NewQueue(&recordingSender{}, email.QueueOptions{})
	t.Cleanup(func() { _ = queue.Shutdown(context.Background()) })

	err := queue.Send(t.Context(), email.Message{To: "nobody", Subject: "Hi", Text: "x"})

	assert.ErrorIs(t, err, email.ErrInvalidMessage)
}
