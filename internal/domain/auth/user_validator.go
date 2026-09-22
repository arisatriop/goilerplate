package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"goilerplate/pkg/constants"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"
)

// dummyPasswordHash is a real bcrypt hash of a value nobody knows. It is compared against when
// there is no password to check — an unregistered email, or a locked account — so that those
// paths take about as long as a genuine comparison. Without it, response time alone would tell
// an attacker which emails are registered.
//
// It is a valid hash at the same cost as utils.HashPassword produces, so the timing matches.
// #nosec G101 -- deliberate: a bcrypt hash of a value nobody knows, existing only to be
// compared against so the no-password path takes as long as a real one. See the comment above.
const dummyPasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

// Lockout describes how many consecutive failures an account tolerates and for how long it is
// then closed to every password.
type Lockout struct {
	MaxAttempts int
	Duration    time.Duration
}

// UserValidator handles user validation logic
type UserValidator struct {
	authRepo Repository
	lockout  Lockout
}

// NewUserValidator creates a new user validator
func NewUserValidator(authRepo Repository, lockout Lockout) *UserValidator {
	return &UserValidator{
		authRepo: authRepo,
		lockout:  lockout,
	}
}

// ValidateUserForLogin checks credentials in an order chosen so that the response reveals as
// little as possible about the account:
//
//  1. Unknown email — answered exactly like a wrong password, after a dummy hash comparison so
//     the timing matches too.
//  2. Locked — rejected without evaluating the password at all. Checking it while locked would
//     let an attacker keep guessing through the lockout and learn from the response when a
//     guess was right, which is the whole thing a lockout exists to prevent.
//  3. Wrong password — counted, possibly locking the account, then answered generically.
//  4. Disabled — reported plainly, but only to someone who already proved they know the
//     password, so it cannot be used to enumerate accounts.
func (uv *UserValidator) ValidateUserForLogin(ctx context.Context, email, password string) (*User, error) {
	user, err := uv.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if user == nil {
		uv.spendPasswordCheckTime(password)
		// No user_id to record, because no account was targeted — the attempted address is
		// deliberately left out of the audit trail. What matters here is the client IP, which
		// is what a burst of these is detected by.
		logger.Security(ctx, logger.SecurityEvent{
			Action:  logger.ActionLoginFailed,
			Outcome: logger.OutcomeFailure,
			Reason:  logger.ReasonUnknownEmail,
		})
		return nil, invalidCredentials()
	}

	// Clear a lock whose window has passed, so the account is usable again on this very attempt
	// rather than only on the next one.
	if user.HasExpiredLock() {
		if err := uv.authRepo.ResetExpiredLock(ctx, user.ID); err != nil {
			return nil, fmt.Errorf("failed to reset expired lock: %w", err)
		}
		user.FailedLoginAttempts = 0
		user.LockedUntil = nil
	}

	if user.IsLocked() {
		uv.spendPasswordCheckTime(password)
		logger.Security(ctx, logger.SecurityEvent{
			Action:  logger.ActionLoginFailed,
			Outcome: logger.OutcomeFailure,
			UserID:  user.ID,
			Reason:  logger.ReasonAccountLocked,
		})
		return nil, utils.ClientErr(http.StatusUnauthorized, constants.MsgAccountLocked)
	}

	if err := utils.CheckPassword(password, user.PasswordHash); err != nil {
		logger.Security(ctx, logger.SecurityEvent{
			Action:  logger.ActionLoginFailed,
			Outcome: logger.OutcomeFailure,
			UserID:  user.ID,
			Reason:  logger.ReasonBadPassword,
		})
		uv.registerFailedLogin(ctx, user.ID)
		return nil, invalidCredentials()
	}

	if !user.IsActive {
		logger.Security(ctx, logger.SecurityEvent{
			Action:  logger.ActionLoginFailed,
			Outcome: logger.OutcomeFailure,
			UserID:  user.ID,
			Reason:  logger.ReasonAccountDisabled,
		})
		return nil, utils.ClientErr(http.StatusForbidden, constants.MsgAccountDisabled)
	}

	return user, nil
}

// ValidateUserForRefresh validates user for refresh token operation
func (uv *UserValidator) ValidateUserForRefresh(ctx context.Context, userID string) (*User, error) {
	user, err := uv.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		return nil, utils.ClientErr(http.StatusNotFound, constants.MsgResourceNotFound)
	}

	if !user.IsActive {
		return nil, utils.ClientErr(http.StatusForbidden, constants.MsgAccountDisabled)
	}

	return user, nil
}

// registerFailedLogin counts the attempt and lets the repository lock the account in the same
// statement. A failure here is logged rather than returned: the caller is being told their
// credentials were wrong either way, and surfacing a database error would turn a failed login
// into a 500 that distinguishes this account from others.
func (uv *UserValidator) registerFailedLogin(ctx context.Context, userID string) {
	lockUntil := utils.Now().Add(uv.lockout.Duration)

	locked, err := uv.authRepo.RegisterFailedLogin(ctx, userID, uv.lockout.MaxAttempts, lockUntil)
	if err != nil {
		logger.Error(ctx, fmt.Errorf("recording failed login for user %s: %w", userID, err))
		return
	}

	if locked {
		logger.Security(ctx, logger.SecurityEvent{
			Action:  logger.ActionAccountLocked,
			Outcome: logger.OutcomeFailure,
			UserID:  userID,
			Reason:  logger.ReasonBadPassword,
		})
		logger.Warn(ctx, fmt.Sprintf(
			"account %s locked until %s after %d failed login attempts",
			userID, lockUntil.Format(time.RFC3339), uv.lockout.MaxAttempts))
	}
}

// spendPasswordCheckTime runs a bcrypt comparison that is certain to fail, so a request with no
// real password to verify costs the same as one that has.
func (uv *UserValidator) spendPasswordCheckTime(password string) {
	_ = utils.CheckPassword(password, dummyPasswordHash)
}

// invalidCredentials is the one answer given for an unknown email and for a wrong password, so
// the two are indistinguishable. 401 rather than 400: the request was well formed, the
// credentials were not accepted.
func invalidCredentials() error {
	return utils.ClientErr(http.StatusUnauthorized, constants.MsgInvalidCredential)
}
