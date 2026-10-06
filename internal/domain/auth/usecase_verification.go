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
	"time"

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

	cooling, err := uc.inOTPCooldown(ctx, user.ID, OneTimeTokenEmailVerification)
	if err != nil {
		return err
	}
	if cooling {
		uc.logVerification(ctx, logger.ActionEmailVerificationRequested, user.ID, logger.OutcomeFailure, logger.ReasonCooldown)
		return nil
	}

	code, expiresAt, err := uc.issueOTP(ctx, user.ID, OneTimeTokenEmailVerification, "", origin)
	if err != nil {
		return err
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
// Every failure is ErrInvalidVerificationCode; matchOTP describes the attempt ceiling.
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

	token, reason, err := uc.matchOTP(ctx, user.ID, OneTimeTokenEmailVerification, code)
	if err != nil {
		return err
	}
	if token == nil {
		return uc.refuseVerification(ctx, user.ID, reason)
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

// matchOTP checks code against the user's current token of tokenType and returns the token when
// it matches. A refusal is a nil token and the reason to log; err is reserved for failures
// that are not the caller's doing.
//
// Each code tolerates OTP.MaxAttempts guesses. The attempt is counted before the code is
// compared, in one atomic UPDATE, so a burst of concurrent guesses cannot all be compared
// against the code before any of them is counted: the ceiling holds however the guesses
// arrive. The guess that uses up the last attempt expires the code.
func (uc *authUseCase) matchOTP(ctx context.Context, userID, tokenType, code string) (*OneTimeToken, string, error) {
	token, err := uc.authRepo.GetLatestActiveOneTimeToken(ctx, userID, tokenType)
	if err != nil {
		return nil, "", fmt.Errorf("getting %s code: %w", tokenType, err)
	}
	if token == nil {
		return nil, logger.ReasonInvalidToken, nil
	}

	attempts, err := uc.authRepo.IncrementOneTimeTokenAttempts(ctx, token.ID)
	if errors.Is(err, ErrNotFound) {
		return nil, logger.ReasonInvalidToken, nil // consumed meanwhile
	}
	if err != nil {
		return nil, "", fmt.Errorf("counting %s attempt: %w", tokenType, err)
	}

	maxAttempts := max(uc.email.OTP.MaxAttempts, 1)
	if attempts > maxAttempts {
		return nil, logger.ReasonTooManyAttempts, nil
	}

	if !hash.KeyedEqual(uc.email.OTP.Secret, otpMessage(token.ID, code), token.TokenHash) {
		if attempts < maxAttempts {
			return nil, logger.ReasonInvalidToken, nil
		}
		// That was the last guess this code allows. Expiring it now, rather than leaving the
		// count to refuse the next one, makes the next request ask for a new code.
		if err := uc.authRepo.ExpireOneTimeTokens(ctx, userID, tokenType); err != nil {
			return nil, "", fmt.Errorf("expiring exhausted %s code: %w", tokenType, err)
		}
		return nil, logger.ReasonTooManyAttempts, nil
	}

	return token, "", nil
}

func (uc *authUseCase) refuseVerification(ctx context.Context, userID, reason string) error {
	uc.logVerification(ctx, logger.ActionEmailVerified, userID, logger.OutcomeFailure, reason)
	return ErrInvalidVerificationCode
}

func (uc *authUseCase) logVerification(ctx context.Context, action, userID, outcome, reason string) {
	logger.Security(ctx, logger.SecurityEvent{Action: action, Outcome: outcome, UserID: userID, Reason: reason})
}

// issueOTP replaces the user's code of tokenType with a new one and returns its plaintext. Only
// the newest code works, so a code that arrived late cannot be used after a new one.
func (uc *authUseCase) issueOTP(ctx context.Context, userID, tokenType, newEmail string, origin RequestOrigin) (string, time.Time, error) {
	code, err := generateOTP()
	if err != nil {
		return "", time.Time{}, err
	}

	// The hash is keyed by the token's own ID as well as the server secret. With the code alone,
	// two users drawing the same code — or one user drawing a code a second time — would store
	// the same hash, which the unique index on token_hash refuses.
	tokenID := utils.GenerateUUID()
	expiresAt := utils.Now().Add(uc.email.OTP.TTL)

	err = uc.txManager.Do(ctx, func(txCtx context.Context) error {
		repo := uc.authRepo.WithTx(txCtx)

		if err := repo.ExpireOneTimeTokens(txCtx, userID, tokenType); err != nil {
			return fmt.Errorf("expiring previous %s codes: %w", tokenType, err)
		}
		return repo.CreateOneTimeToken(txCtx, &OneTimeToken{
			ID:        tokenID,
			UserID:    userID,
			TokenType: tokenType,
			TokenHash: uc.otpHash(tokenID, code),
			ExpiresAt: expiresAt,
			IPAddress: origin.IPAddress,
			UserAgent: origin.UserAgent,
			NewEmail:  newEmail,
		})
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("issuing %s code: %w", tokenType, err)
	}
	return code, expiresAt, nil
}

// inOTPCooldown reports whether the user was sent a code of tokenType too recently for another.
func (uc *authUseCase) inOTPCooldown(ctx context.Context, userID, tokenType string) (bool, error) {
	latest, err := uc.authRepo.GetLatestActiveOneTimeToken(ctx, userID, tokenType)
	if err != nil {
		return false, fmt.Errorf("getting latest %s code: %w", tokenType, err)
	}
	return latest != nil && utils.Now().Sub(latest.CreatedAt) < uc.email.OTP.ResendCooldown, nil
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
