// Package job holds the background work the application runs for itself.
package job

import (
	"context"
	"fmt"
	"time"

	"goilerplate/internal/domain/lock"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"
)

// cleanupLockKey names the lock. One cluster, one cleanup at a time.
const cleanupLockKey = "job:cleanup"

// maxBatchesPerRun bounds a single run. Without it, a first run against a table with years of
// history would hold the job for as long as the backlog took, and a bug in the delete predicate
// would have no natural stopping point. Whatever is left is picked up on the next tick.
const maxBatchesPerRun = 100

// CleanupRepository deletes rows that can no longer be used, oldest first, at most batchSize at
// a time. Each call reports how many it removed, so the caller can tell a full batch (there may
// be more) from a partial one (there is not).
type CleanupRepository interface {
	DeleteExpiredSessions(ctx context.Context, cutoff time.Time, batchSize int) (int64, error)
	DeleteFinishedOneTimeTokens(ctx context.Context, cutoff time.Time, batchSize int) (int64, error)
}

// Result reports what one run removed.
type Result struct {
	Sessions       int64
	OneTimeTokens  int64
	SessionBatches int
	TokenBatches   int
}

// Cleanup deletes revoked or expired sessions and finished one-time tokens once they are older
// than the retention period.
//
// The retention is the point: a revoked session is the record of a logout, and an incident is
// usually investigated well after it happened. Rows are removed because the tables would
// otherwise grow without bound, not because the history stops being useful the moment a session
// ends.
type Cleanup struct {
	repo      CleanupRepository
	locker    lock.Provider
	interval  time.Duration
	retention time.Duration
	batchSize int
}

func NewCleanup(repo CleanupRepository, locker lock.Provider, interval, retention time.Duration, batchSize int) *Cleanup {
	return &Cleanup{
		repo:      repo,
		locker:    locker,
		interval:  interval,
		retention: retention,
		batchSize: batchSize,
	}
}

// Run ticks until ctx is cancelled. It does not run at startup: a deploy or a crash loop would
// otherwise turn every restart into a delete pass.
func (c *Cleanup) Run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	logger.Info(ctx, fmt.Sprintf("cleanup job started: every %s, retaining %s", c.interval, c.retention))

	for {
		select {
		case <-ctx.Done():
			logger.Info(ctx, "cleanup job stopped")
			return
		case <-ticker.C:
			c.runLocked(ctx)
		}
	}
}

// runLocked performs one pass if no other instance is already doing it.
//
// The lock is an optimisation, not a correctness requirement: the deletes are idempotent, so two
// instances overlapping would waste work rather than break anything. That is what makes it safe
// for the lock to expire under a run that takes longer than expected.
func (c *Cleanup) runLocked(ctx context.Context) {
	release, acquired, err := c.locker.TryLock(ctx, cleanupLockKey, c.interval)
	if err != nil {
		logger.Error(ctx, fmt.Errorf("acquiring cleanup lock: %w", err))
		return
	}
	if !acquired {
		// Expected on every instance but one. Not worth an INFO line every interval.
		logger.Debug(ctx, "cleanup job skipped: another instance holds the lock")
		return
	}
	defer func() {
		// WithoutCancel so a shutdown mid-run still releases the lock rather than leaving it
		// held until its TTL expires.
		if err := release(context.WithoutCancel(ctx)); err != nil {
			logger.Error(ctx, fmt.Errorf("releasing cleanup lock: %w", err))
		}
	}()

	result, err := c.RunOnce(ctx)
	if err != nil {
		logger.Error(ctx, fmt.Errorf("cleanup job: %w", err))
		return
	}

	if result.Sessions > 0 || result.OneTimeTokens > 0 {
		logger.Info(ctx, fmt.Sprintf("cleanup removed %d sessions and %d one-time tokens",
			result.Sessions, result.OneTimeTokens))
	}
}

// RunOnce performs one pass and reports what it removed. Exported so a deployment can trigger a
// pass from its own scheduler instead of the built-in ticker.
func (c *Cleanup) RunOnce(ctx context.Context) (Result, error) {
	cutoff := utils.Now().Add(-c.retention)
	var result Result
	var err error

	result.Sessions, result.SessionBatches, err = c.deleteInBatches(ctx, cutoff, c.repo.DeleteExpiredSessions)
	if err != nil {
		return result, fmt.Errorf("deleting expired sessions: %w", err)
	}

	result.OneTimeTokens, result.TokenBatches, err = c.deleteInBatches(ctx, cutoff, c.repo.DeleteFinishedOneTimeTokens)
	if err != nil {
		return result, fmt.Errorf("deleting finished one-time tokens: %w", err)
	}

	return result, nil
}

// deleteInBatches calls deleteBatch until it removes less than a full batch, meaning nothing
// eligible is left. Cancellation is honoured between batches, so a shutdown stops after the
// batch in flight rather than abandoning a run halfway through a statement.
func (c *Cleanup) deleteInBatches(
	ctx context.Context,
	cutoff time.Time,
	deleteBatch func(context.Context, time.Time, int) (int64, error),
) (int64, int, error) {
	var total int64
	var batches int

	for batches < maxBatchesPerRun {
		if err := ctx.Err(); err != nil {
			return total, batches, nil //nolint:nilerr // a cancelled run is not a failed one
		}

		removed, err := deleteBatch(ctx, cutoff, c.batchSize)
		if err != nil {
			return total, batches, err
		}

		total += removed
		batches++

		if removed < int64(c.batchSize) {
			break
		}
	}

	return total, batches, nil
}
