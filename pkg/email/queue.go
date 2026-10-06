package email

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
)

// DefaultQueueSize is how many messages a Queue holds before Send starts refusing.
const DefaultQueueSize = 256

// Errors returned by Queue.Send.
var (
	ErrQueueFull   = errors.New("email: queue is full")
	ErrQueueClosed = errors.New("email: queue is shut down")
)

// Queue delivers messages on one background worker. Send returns as soon as the message is
// queued, so the request that triggered an email neither waits for the provider nor reveals by
// its latency whether an email was sent at all — which, for "forgot password", would say
// whether the address has an account.
//
// The queue is in memory: a message still queued when the process dies is lost. That suits
// mail a user can ask for again; anything that must survive a crash needs an outbox table.
//
// The worker starts in NewQueue. Its owner must call Shutdown.
type Queue struct {
	sender Sender
	logger *slog.Logger

	mu     sync.RWMutex // guards closed against a Send racing Shutdown's close(jobs)
	closed bool
	jobs   chan Message

	// base parents every delivery. It is cancelled only when Shutdown runs out of time, so a
	// delivery in flight at shutdown is allowed to finish rather than cut off mid-DATA.
	base    context.Context
	abandon context.CancelFunc
	done    chan struct{}
	dropped atomic.Int64
}

// QueueOptions configures a Queue. The zero value is usable.
type QueueOptions struct {
	// Size is the queue capacity. Default DefaultQueueSize.
	Size int
	// Logger receives delivery failures. Default slog.Default.
	Logger *slog.Logger
}

// NewQueue starts a worker delivering through sender.
func NewQueue(sender Sender, opts QueueOptions) *Queue {
	size := opts.Size
	if size <= 0 {
		size = DefaultQueueSize
	}
	logger := opts.Logger
	if logger == nil {
		logger = slog.Default()
	}

	base, abandon := context.WithCancel(context.Background())
	q := &Queue{
		sender:  sender,
		logger:  logger,
		jobs:    make(chan Message, size),
		base:    base,
		abandon: abandon,
		done:    make(chan struct{}),
	}
	go q.run()
	return q
}

// Send queues msg for delivery. It never blocks: a full queue is reported rather than waited
// on, so a stalled provider cannot back up into request handling. A malformed message is
// refused here, where the caller can still see the error.
func (q *Queue) Send(_ context.Context, msg Message) error {
	if _, err := msg.validate(); err != nil {
		return err
	}

	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.closed {
		return ErrQueueClosed
	}
	select {
	case q.jobs <- msg:
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *Queue) run() {
	defer close(q.done)

	for msg := range q.jobs {
		if q.base.Err() != nil {
			q.dropped.Add(1)
			continue
		}

		// The worker is the boundary for these errors: nobody is waiting on the result.
		// The recipient is left out of the log line on purpose; it is personal data and the
		// subject is enough to tell which flow failed.
		if err := q.sender.Send(q.base, msg); err != nil {
			if q.base.Err() != nil {
				q.dropped.Add(1) // abandoned by Shutdown, which reports the count
				continue
			}
			q.logger.Error("email delivery failed", "subject", msg.Subject, "error", err)
		}
	}
}

// Shutdown stops accepting messages and waits for the queued ones to be delivered. When ctx
// ends first, the delivery in flight is cancelled, the rest are dropped, and the returned error
// says how many were lost. Calling it again is safe.
func (q *Queue) Shutdown(ctx context.Context) error {
	q.mu.Lock()
	if !q.closed {
		q.closed = true
		close(q.jobs)
	}
	q.mu.Unlock()

	select {
	case <-q.done:
		return nil
	case <-ctx.Done():
		q.abandon()
		<-q.done
		return fmt.Errorf("email queue: %d message(s) not delivered: %w", q.dropped.Load(), ctx.Err())
	}
}
