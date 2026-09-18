package repository

import (
	"context"
	"time"

	"goilerplate/internal/domain/job"
	"goilerplate/pkg/utils"

	"gorm.io/gorm"
)

type cleanupRepository struct {
	db *gorm.DB
}

func NewCleanup(db *gorm.DB) job.CleanupRepository {
	return &cleanupRepository{db: db}
}

// DeleteExpiredSessions removes sessions that can no longer be used and stopped being usable
// before cutoff.
//
// Two conditions, deliberately, where one would do. `COALESCE(revoked_at, expires_at) < cutoff`
// already implies the row is dead whenever cutoff is in the past — but this statement deletes
// login history, and a misconfigured retention is the one input that could put cutoff in the
// future. The explicit "revoked, or expired by now" test means that mistake deletes nothing
// instead of deleting every live session.
//
// PostgreSQL has no DELETE ... LIMIT, so the batch is chosen by a subquery on the primary key.
func (r *cleanupRepository) DeleteExpiredSessions(ctx context.Context, cutoff time.Time, batchSize int) (int64, error) {
	now := utils.Now()

	result := r.db.WithContext(ctx).Exec(`
		DELETE FROM user_sessions
		 WHERE id IN (
		       SELECT id FROM user_sessions
		        WHERE (is_active = false OR expires_at <= ?)
		          AND COALESCE(revoked_at, expires_at) < ?
		        ORDER BY COALESCE(revoked_at, expires_at)
		        LIMIT ?
		 )`, now, cutoff, batchSize)

	return result.RowsAffected, result.Error
}

// DeleteFinishedOneTimeTokens removes one-time tokens that were consumed or have expired, and
// reached that state before cutoff. Same two-condition reasoning as above: an unused token that
// has not expired is still a pending password reset.
func (r *cleanupRepository) DeleteFinishedOneTimeTokens(ctx context.Context, cutoff time.Time, batchSize int) (int64, error) {
	now := utils.Now()

	result := r.db.WithContext(ctx).Exec(`
		DELETE FROM one_time_tokens
		 WHERE id IN (
		       SELECT id FROM one_time_tokens
		        WHERE (used_at IS NOT NULL OR expires_at <= ?)
		          AND COALESCE(used_at, expires_at) < ?
		        ORDER BY COALESCE(used_at, expires_at)
		        LIMIT ?
		 )`, now, cutoff, batchSize)

	return result.RowsAffected, result.Error
}
