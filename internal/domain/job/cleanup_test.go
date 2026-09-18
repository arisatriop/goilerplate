package job

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"goilerplate/internal/domain/lock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// countingRepo reports a fixed backlog and removes up to batchSize of it per call, which is what
// a real batched DELETE does.
type countingRepo struct {
	mu             sync.Mutex
	sessions       int64
	tokens         int64
	sessionCalls   int
	tokenCalls     int
	cutoffs        []time.Time
	sessionErr     error
	lastBatchSizes []int
}

func (r *countingRepo) DeleteExpiredSessions(_ context.Context, cutoff time.Time, batchSize int) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.sessionCalls++
	r.cutoffs = append(r.cutoffs, cutoff)
	r.lastBatchSizes = append(r.lastBatchSizes, batchSize)
	if r.sessionErr != nil {
		return 0, r.sessionErr
	}
	return takeUpTo(&r.sessions, batchSize), nil
}

func (r *countingRepo) DeleteFinishedOneTimeTokens(_ context.Context, _ time.Time, batchSize int) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tokenCalls++
	return takeUpTo(&r.tokens, batchSize), nil
}

func takeUpTo(remaining *int64, batchSize int) int64 {
	removed := int64(batchSize)
	if *remaining < removed {
		removed = *remaining
	}
	*remaining -= removed
	return removed
}

// heldLock never grants the lock, standing in for another instance already running a pass.
type heldLock struct{}

func (heldLock) TryLock(context.Context, string, time.Duration) (lock.Release, bool, error) {
	return nil, false, nil
}

// freeLock always grants, and records that the release was called.
type freeLock struct{ released bool }

func (l *freeLock) TryLock(context.Context, string, time.Duration) (lock.Release, bool, error) {
	return func(context.Context) error { l.released = true; return nil }, true, nil
}

// failingLock stands in for Redis being unreachable.
type failingLock struct{}

func (failingLock) TryLock(context.Context, string, time.Duration) (lock.Release, bool, error) {
	return nil, false, errors.New("redis unavailable")
}

func TestRunOnce_DeletesInBatchesUntilNothingIsLeft(t *testing.T) {
	repo := &countingRepo{sessions: 250, tokens: 40}
	cleanup := NewCleanup(repo, &freeLock{}, time.Hour, 24*time.Hour, 100)

	result, err := cleanup.RunOnce(context.Background())

	require.NoError(t, err)
	assert.Equal(t, int64(250), result.Sessions)
	assert.Equal(t, int64(40), result.OneTimeTokens)

	// 100, 100, 50 — the short third batch is what says the backlog is exhausted.
	assert.Equal(t, 3, result.SessionBatches)
	assert.Equal(t, 3, repo.sessionCalls)

	// One partial batch is enough on its own.
	assert.Equal(t, 1, result.TokenBatches)
	assert.Equal(t, 1, repo.tokenCalls)
}

// An exactly-full final batch cannot be distinguished from "there may be more", so the job asks
// once more and gets an empty batch. Stopping on a full batch would leave rows behind forever.
func TestRunOnce_ExactMultipleOfBatchSizeAsksOnceMore(t *testing.T) {
	repo := &countingRepo{sessions: 200}
	cleanup := NewCleanup(repo, &freeLock{}, time.Hour, 24*time.Hour, 100)

	result, err := cleanup.RunOnce(context.Background())

	require.NoError(t, err)
	assert.Equal(t, int64(200), result.Sessions)
	assert.Equal(t, 3, result.SessionBatches, "100, 100, then an empty batch to confirm")
}

// Without a ceiling, a first pass over years of history would run for as long as the backlog
// took, and a bug in the delete predicate would have no natural stopping point.
func TestRunOnce_StopsAtTheBatchCeiling(t *testing.T) {
	repo := &countingRepo{sessions: 1_000_000}
	cleanup := NewCleanup(repo, &freeLock{}, time.Hour, 24*time.Hour, 10)

	result, err := cleanup.RunOnce(context.Background())

	require.NoError(t, err)
	assert.Equal(t, maxBatchesPerRun, result.SessionBatches)
	assert.Equal(t, int64(maxBatchesPerRun*10), result.Sessions)
	assert.Positive(t, repo.sessions, "the remainder waits for the next tick")
}

// The cutoff is what protects live data, so it must be retention behind now — never ahead of it.
func TestRunOnce_CutoffIsRetentionInThePast(t *testing.T) {
	repo := &countingRepo{}
	retention := 30 * 24 * time.Hour
	cleanup := NewCleanup(repo, &freeLock{}, time.Hour, retention, 100)

	before := time.Now()
	_, err := cleanup.RunOnce(context.Background())
	require.NoError(t, err)

	require.NotEmpty(t, repo.cutoffs)
	cutoff := repo.cutoffs[0]
	assert.True(t, cutoff.Before(before), "cutoff must be in the past")
	assert.WithinDuration(t, before.Add(-retention), cutoff, time.Minute)
}

func TestRunOnce_ReturnsRepositoryErrors(t *testing.T) {
	repo := &countingRepo{sessions: 10, sessionErr: errors.New("connection refused")}
	cleanup := NewCleanup(repo, &freeLock{}, time.Hour, 24*time.Hour, 100)

	_, err := cleanup.RunOnce(context.Background())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "deleting expired sessions")
	assert.Zero(t, repo.tokenCalls, "the pass stops rather than carrying on to the next table")
}

// A shutdown mid-pass stops after the batch in flight. What was already deleted stays deleted;
// the rest is picked up next time.
func TestRunOnce_StopsWhenTheContextIsCancelled(t *testing.T) {
	repo := &countingRepo{sessions: 1000}
	cleanup := NewCleanup(repo, &freeLock{}, time.Hour, 24*time.Hour, 10)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := cleanup.RunOnce(ctx)

	require.NoError(t, err, "a cancelled run is not a failed one")
	assert.Zero(t, result.SessionBatches)
	assert.Zero(t, repo.sessionCalls)
}

// With Redis, every instance but one finds the lock held and does nothing.
func TestRunLocked_SkipsWhenAnotherInstanceHoldsTheLock(t *testing.T) {
	repo := &countingRepo{sessions: 100}
	cleanup := NewCleanup(repo, heldLock{}, time.Hour, 24*time.Hour, 10)

	cleanup.runLocked(context.Background())

	assert.Zero(t, repo.sessionCalls, "no deletes at all, not merely a shorter pass")
}

func TestRunLocked_ReleasesTheLockAfterAPass(t *testing.T) {
	repo := &countingRepo{sessions: 5}
	locker := &freeLock{}
	cleanup := NewCleanup(repo, locker, time.Hour, 24*time.Hour, 10)

	cleanup.runLocked(context.Background())

	assert.Positive(t, repo.sessionCalls)
	assert.True(t, locker.released, "a held lock would stall every instance until its TTL expired")
}

// Redis being unreachable must not turn into a pass that runs everywhere at once.
func TestRunLocked_DoesNotRunWhenTheLockCannotBeAcquired(t *testing.T) {
	repo := &countingRepo{sessions: 100}
	cleanup := NewCleanup(repo, failingLock{}, time.Hour, 24*time.Hour, 10)

	cleanup.runLocked(context.Background())

	assert.Zero(t, repo.sessionCalls)
}

// Startup must not be a delete pass: a crash loop would otherwise hammer the database.
func TestRun_DoesNotRunAtStartupAndStopsOnCancel(t *testing.T) {
	repo := &countingRepo{sessions: 100}
	cleanup := NewCleanup(repo, &freeLock{}, time.Hour, 24*time.Hour, 10)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { cleanup.Run(ctx); close(done) }()

	time.Sleep(50 * time.Millisecond)
	repo.mu.Lock()
	callsBeforeFirstTick := repo.sessionCalls
	repo.mu.Unlock()

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after its context was cancelled")
	}

	assert.Zero(t, callsBeforeFirstTick, "the first pass waits for a tick")
}

func TestRun_TicksAtTheConfiguredInterval(t *testing.T) {
	repo := &countingRepo{sessions: 0}
	cleanup := NewCleanup(repo, &freeLock{}, 20*time.Millisecond, 24*time.Hour, 10)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	cleanup.Run(ctx)

	repo.mu.Lock()
	defer repo.mu.Unlock()
	assert.Positive(t, repo.sessionCalls, "the ticker must actually fire")
}
