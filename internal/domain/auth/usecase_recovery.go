// Account recovery: "forgot password" and the reset that follows. Both answer without revealing
// whether an address has an account, and a reset turns out every session, because the user
// asking for one may be doing it precisely because someone else got in.

package auth

import (
	"context"
	"errors"
	"fmt"

	"goilerplate/pkg/hash"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"
)

// resetTokenBytes is the entropy of a reset token: 256 bits, far beyond guessing, which is
// what lets it be stored as a plain SHA-256 rather than a keyed hash (see OneTimeToken).
const resetTokenBytes = 32

// errEmailFlowsUnavailable is a wiring fault, not a client error: the routes are registered only
// when email is configured, so reaching this means the two disagree.
var errEmailFlowsUnavailable = errors.New("auth: email flows are not configured")

// ForgotPassword emails a reset link to the account registered under email, if there is one.
//
// Every outcome returns nil — unknown address, disabled account, a request inside the cooldown,
// even a failure to queue the email — so the response cannot be used to discover which
// addresses are registered. The difference is visible only in the security log.
func (uc *authUseCase) ForgotPassword(ctx context.Context, email string, origin RequestOrigin) error {
	if uc.email.Notifier == nil {
		return errEmailFlowsUnavailable
	}

	user, err := uc.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}
	if user == nil {
		uc.logResetRequested(ctx, "", logger.OutcomeFailure, logger.ReasonUnknownEmail)
		return nil
	}
	if !user.IsActive {
		uc.logResetRequested(ctx, user.ID, logger.OutcomeFailure, logger.ReasonAccountDisabled)
		return nil
	}

	latest, err := uc.authRepo.GetLatestActiveOneTimeToken(ctx, user.ID, OneTimeTokenPasswordReset)
	if err != nil {
		return fmt.Errorf("getting latest reset token: %w", err)
	}
	if latest != nil && utils.Now().Sub(latest.CreatedAt) < uc.email.ResetResendCooldown {
		uc.logResetRequested(ctx, user.ID, logger.OutcomeFailure, logger.ReasonCooldown)
		return nil
	}

	token, err := utils.GenerateSecureToken(resetTokenBytes)
	if err != nil {
		return fmt.Errorf("generating reset token: %w", err)
	}
	expiresAt := utils.Now().Add(uc.email.ResetTTL)

	// Only the newest link works. Expiring the older ones in the same transaction means a link
	// sitting in an inbox that has since been compromised stops working the moment a new one
	// is asked for.
	err = uc.txManager.Do(ctx, func(txCtx context.Context) error {
		repo := uc.authRepo.WithTx(txCtx)

		if err := repo.ExpireOneTimeTokens(txCtx, user.ID, OneTimeTokenPasswordReset); err != nil {
			return fmt.Errorf("expiring previous reset tokens: %w", err)
		}
		return repo.CreateOneTimeToken(txCtx, &OneTimeToken{
			UserID:    user.ID,
			TokenType: OneTimeTokenPasswordReset,
			TokenHash: hash.Token(token),
			ExpiresAt: expiresAt,
			IPAddress: origin.IPAddress,
			UserAgent: origin.UserAgent,
		})
	})
	if err != nil {
		return fmt.Errorf("issuing reset token: %w", err)
	}

	notice := PasswordResetNotice{Email: user.Email, Name: user.Name, Token: token, ExpiresAt: expiresAt}
	if err := uc.email.Notifier.SendPasswordReset(ctx, notice); err != nil {
		// Handled here rather than returned: a 500 for a registered address and a 200 for an
		// unknown one is exactly the difference this endpoint must not show. The user can ask
		// again; the operator learns of it from this line.
		logger.Error(ctx, fmt.Errorf("queueing password reset email: %w", err))
		uc.logResetRequested(ctx, user.ID, logger.OutcomeFailure, logger.ReasonDeliveryFailed)
		return nil
	}

	uc.logResetRequested(ctx, user.ID, logger.OutcomeSuccess, "")
	return nil
}

// ResetPassword sets a new password for the account the token was issued to, and revokes every
// session that account has.
//
// The new password is checked before the token is touched, so a password the policy refuses
// does not use up the link.
func (uc *authUseCase) ResetPassword(ctx context.Context, token, newPassword string) error {
	if uc.email.Notifier == nil {
		return errEmailFlowsUnavailable
	}
	if token == "" {
		return ErrInvalidResetToken
	}

	if err := uc.passwordPolicy.Validate(newPassword); err != nil {
		return err
	}
	passwordHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hashing new password: %w", err)
	}

	// Consuming the token, replacing the password and revoking the sessions commit together.
	// Consumption is a single conditional UPDATE, so two requests racing with the same link
	// yield one reset and one ErrInvalidResetToken.
	var userID string
	err = uc.txManager.Do(ctx, func(txCtx context.Context) error {
		repo := uc.authRepo.WithTx(txCtx)

		consumed, err := repo.ConsumeOneTimeToken(txCtx, hash.Token(token), OneTimeTokenPasswordReset)
		if errors.Is(err, ErrNotFound) {
			return ErrInvalidResetToken
		}
		if err != nil {
			return fmt.Errorf("consuming reset token: %w", err)
		}

		user, err := repo.GetUserByID(txCtx, consumed.UserID)
		if err != nil {
			return fmt.Errorf("getting user: %w", err)
		}
		// An account disabled after the link was sent must stay disabled. Rolling back also
		// leaves the token unused, which is harmless: it can never succeed while this holds.
		if user == nil || !user.IsActive {
			return ErrInvalidResetToken
		}

		if err := repo.UpdateUserPassword(txCtx, user.ID, passwordHash); err != nil {
			return fmt.Errorf("updating password: %w", err)
		}
		if err := repo.RevokeOtherUserSessions(txCtx, user.ID, "", RevokedReasonPasswordReset); err != nil {
			return fmt.Errorf("revoking sessions: %w", err)
		}
		// Following the emailed link proves the user reads that inbox, which is all
		// verification asks. Without this, a user who reset before verifying would be locked
		// out again by auth.require_email_verification right after choosing a new password.
		if !user.EmailVerified {
			if err := repo.MarkEmailVerified(txCtx, user.ID); err != nil {
				return fmt.Errorf("marking email verified: %w", err)
			}
		}

		userID = user.ID
		return nil
	})
	if errors.Is(err, ErrInvalidResetToken) {
		logger.Security(ctx, logger.SecurityEvent{
			Action:  logger.ActionPasswordResetCompleted,
			Outcome: logger.OutcomeFailure,
			Reason:  logger.ReasonInvalidToken,
		})
		return err
	}
	if err != nil {
		return err
	}

	uc.sessionService.EvictUser(ctx, userID)

	logger.Security(ctx, logger.SecurityEvent{
		Action:  logger.ActionPasswordResetCompleted,
		Outcome: logger.OutcomeSuccess,
		UserID:  userID,
	})
	return nil
}

func (uc *authUseCase) logResetRequested(ctx context.Context, userID, outcome, reason string) {
	logger.Security(ctx, logger.SecurityEvent{
		Action:  logger.ActionPasswordResetRequested,
		Outcome: outcome,
		UserID:  userID,
		Reason:  reason,
	})
}
