// Email verification: a 6-digit code sent to the address, typed back to prove the user reads it.
// Both steps work without a session, because with auth.require_email_verification on, an
// unverified user cannot sign in to get one.

package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"

	"goilerplate/pkg/hash"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"
)

// otpDigits is the length of a verification code. Six digits is what people expect to type;
// the attempt ceiling, not the length, is what makes guessing it impractical.
const otpDigits = 6

var otpSpace = big.NewInt(1_000_000)

// SendEmailVerification emails a verification code to the account registered under email.
//
// Like ForgotPassword, it returns nil whatever happens — unknown address, disabled account,
// already verified, inside the cooldown, or a failure to queue the email — so it cannot be used
// to discover which addresses are registered or verified.
func (uc *authUseCase) SendEmailVerification(ctx context.Context, email string, origin RequestOrigin) error {
	if uc.email.Notifier == nil {
		return errEmailFlowsUnavailable
	}

	user, err := uc.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}
	switch {
	case user == nil:
		uc.logVerification(ctx, logger.ActionEmailVerificationRequested, "", logger.OutcomeFailure, logger.ReasonUnknownEmail)
		return nil
	case !user.IsActive:
		uc.logVerification(ctx, logger.ActionEmailVerificationRequested, user.ID, logger.OutcomeFailure, logger.ReasonAccountDisabled)
		return nil
	case user.EmailVerified:
		uc.logVerification(ctx, logger.ActionEmailVerificationRequested, user.ID, logger.OutcomeFailure, logger.ReasonAlreadyVerified)
		return nil
	}

	latest, err := uc.authRepo.GetLatestActiveOneTimeToken(ctx, user.ID, OneTimeTokenEmailVerification)
	if err != nil {
		return fmt.Errorf("getting latest verification code: %w", err)
	}
	if latest != nil && utils.Now().Sub(latest.CreatedAt) < uc.email.OTP.ResendCooldown {
		uc.logVerification(ctx, logger.ActionEmailVerificationRequested, user.ID, logger.OutcomeFailure, logger.ReasonCooldown)
		return nil
	}

	code, err := generateOTP()
	if err != nil {
		return err
	}

	// The hash is keyed by the token's own ID as well as the server secret. With the code alone,
	// two users drawing the same code — or one user drawing a code a second time — would store
	// the same hash, which the unique index on token_hash refuses.
	tokenID := utils.GenerateUUID()
	expiresAt := utils.Now().Add(uc.email.OTP.TTL)

	// Only the newest code works, so a code that arrived late cannot be used after a new one.
	err = uc.txManager.Do(ctx, func(txCtx context.Context) error {
		repo := uc.authRepo.WithTx(txCtx)

		if err := repo.ExpireOneTimeTokens(txCtx, user.ID, OneTimeTokenEmailVerification); err != nil {
			return fmt.Errorf("expiring previous verification codes: %w", err)
		}
		return repo.CreateOneTimeToken(txCtx, &OneTimeToken{
			ID:        tokenID,
			UserID:    user.ID,
			TokenType: OneTimeTokenEmailVerification,
			TokenHash: uc.otpHash(tokenID, code),
			ExpiresAt: expiresAt,
			IPAddress: origin.IPAddress,
			UserAgent: origin.UserAgent,
		})
	})
	if err != nil {
		return fmt.Errorf("issuing verification code: %w", err)
	}

	notice := EmailVerificationNotice{Email: user.Email, Name: user.Name, Code: code, ExpiresAt: expiresAt}
	if err := uc.email.Notifier.SendEmailVerification(ctx, notice); err != nil {
		// Handled, not returned, for the same reason as in ForgotPassword.
		logger.Error(ctx, fmt.Errorf("queueing verification email: %w", err))
		uc.logVerification(ctx, logger.ActionEmailVerificationRequested, user.ID, logger.OutcomeFailure, logger.ReasonDeliveryFailed)
		return nil
	}

	uc.logVerification(ctx, logger.ActionEmailVerificationRequested, user.ID, logger.OutcomeSuccess, "")
	return nil
}

// VerifyEmail marks the address verified when code is the account's current verification code.
//
// Each code tolerates OTP.MaxAttempts guesses. The attempt is counted before the code is
// compared, in one atomic UPDATE, so a burst of concurrent guesses cannot all be compared
// against the code before any of them is counted: the ceiling holds however the guesses
// arrive. Every failure is ErrInvalidVerificationCode.
func (uc *authUseCase) VerifyEmail(ctx context.Context, email, code string) error {
	if uc.email.Notifier == nil {
		return errEmailFlowsUnavailable
	}

	user, err := uc.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}
	if user == nil || !user.IsActive {
		return uc.refuseVerification(ctx, "", logger.ReasonInvalidToken)
	}

	token, err := uc.authRepo.GetLatestActiveOneTimeToken(ctx, user.ID, OneTimeTokenEmailVerification)
	if err != nil {
		return fmt.Errorf("getting verification code: %w", err)
	}
	if token == nil {
		return uc.refuseVerification(ctx, user.ID, logger.ReasonInvalidToken)
	}

	attempts, err := uc.authRepo.IncrementOneTimeTokenAttempts(ctx, token.ID)
	if errors.Is(err, ErrNotFound) {
		return uc.refuseVerification(ctx, user.ID, logger.ReasonInvalidToken) // consumed meanwhile
	}
	if err != nil {
		return fmt.Errorf("counting verification attempt: %w", err)
	}

	maxAttempts := max(uc.email.OTP.MaxAttempts, 1)
	if attempts > maxAttempts {
		return uc.refuseVerification(ctx, user.ID, logger.ReasonTooManyAttempts)
	}

	if !hash.KeyedEqual(uc.email.OTP.Secret, otpMessage(token.ID, code), token.TokenHash) {
		if attempts == maxAttempts {
			// That was the last guess this code allows. Expiring it now, rather than leaving
			// the count to refuse the next one, makes the next request ask for a new code.
			if err := uc.authRepo.ExpireOneTimeTokens(ctx, user.ID, OneTimeTokenEmailVerification); err != nil {
				return fmt.Errorf("expiring exhausted verification code: %w", err)
			}
			return uc.refuseVerification(ctx, user.ID, logger.ReasonTooManyAttempts)
		}
		return uc.refuseVerification(ctx, user.ID, logger.ReasonInvalidToken)
	}

	err = uc.txManager.Do(ctx, func(txCtx context.Context) error {
		repo := uc.authRepo.WithTx(txCtx)

		if _, err := repo.ConsumeOneTimeToken(txCtx, token.TokenHash, OneTimeTokenEmailVerification); err != nil {
			if errors.Is(err, ErrNotFound) {
				return ErrInvalidVerificationCode // a concurrent request with the same code won
			}
			return fmt.Errorf("consuming verification code: %w", err)
		}
		if err := repo.MarkEmailVerified(txCtx, user.ID); err != nil {
			return fmt.Errorf("marking email verified: %w", err)
		}
		return nil
	})
	if errors.Is(err, ErrInvalidVerificationCode) {
		return uc.refuseVerification(ctx, user.ID, logger.ReasonInvalidToken)
	}
	if err != nil {
		return err
	}

	uc.logVerification(ctx, logger.ActionEmailVerified, user.ID, logger.OutcomeSuccess, "")
	return nil
}

func (uc *authUseCase) refuseVerification(ctx context.Context, userID, reason string) error {
	uc.logVerification(ctx, logger.ActionEmailVerified, userID, logger.OutcomeFailure, reason)
	return ErrInvalidVerificationCode
}

func (uc *authUseCase) logVerification(ctx context.Context, action, userID, outcome, reason string) {
	logger.Security(ctx, logger.SecurityEvent{Action: action, Outcome: outcome, UserID: userID, Reason: reason})
}

func (uc *authUseCase) otpHash(tokenID, code string) string {
	return hash.Keyed(uc.email.OTP.Secret, otpMessage(tokenID, code))
}

func otpMessage(tokenID, code string) string {
	return tokenID + ":" + code
}

// generateOTP draws a uniformly random 6-digit code, leading zeros included.
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, otpSpace)
	if err != nil {
		return "", fmt.Errorf("generating verification code: %w", err)
	}
	return fmt.Sprintf("%0*d", otpDigits, n.Int64()), nil
}
