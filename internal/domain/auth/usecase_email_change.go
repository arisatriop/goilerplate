// Email change: moving an account to another address. The current password authorises it, a
// code sent to the new address proves the user reads it, and the old address is told once it
// is done.

package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"
)

// RequestEmailChange sends a confirmation code to newEmail. Nothing changes until the code is
// confirmed.
//
// An address another account already holds is answered exactly like a free one, but no code is
// sent: otherwise any signed-in user could probe which addresses are registered. The owner of
// the address can still find out by asking for a reset, which goes to the inbox they control.
func (uc *authUseCase) RequestEmailChange(ctx context.Context, userID, currentPassword, newEmail string, origin RequestOrigin) error {
	if uc.email.Notifier == nil {
		return errEmailFlowsUnavailable
	}

	user, err := uc.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}
	if user == nil || !user.IsActive {
		return ErrUnauthorized
	}

	// A signed-in session is enough to read the account, not to hand it to another inbox — the
	// same reasoning as ChangePassword.
	if err := utils.CheckPassword(currentPassword, user.PasswordHash); err != nil {
		uc.logEmailChange(ctx, userID, logger.OutcomeFailure, logger.ReasonBadPassword)
		return ErrInvalidCredentials
	}
	if strings.EqualFold(strings.TrimSpace(newEmail), user.Email) {
		return ErrEmailUnchanged
	}

	holder, err := uc.authRepo.GetUserByEmail(ctx, newEmail)
	if err != nil {
		return fmt.Errorf("checking new address: %w", err)
	}
	if holder != nil {
		uc.logEmailChange(ctx, userID, logger.OutcomeFailure, logger.ReasonAddressTaken)
		return nil
	}

	cooling, err := uc.inOTPCooldown(ctx, userID, OneTimeTokenEmailChange)
	if err != nil {
		return err
	}
	if cooling {
		uc.logEmailChange(ctx, userID, logger.OutcomeFailure, logger.ReasonCooldown)
		return nil
	}

	code, expiresAt, err := uc.issueOTP(ctx, userID, OneTimeTokenEmailChange, newEmail, origin)
	if err != nil {
		return err
	}

	notice := EmailChangeNotice{NewEmail: newEmail, Name: user.Name, Code: code, ExpiresAt: expiresAt}
	if err := uc.email.Notifier.SendEmailChangeCode(ctx, notice); err != nil {
		// The same answer as a taken address, so a failure here reveals nothing either.
		logger.Error(ctx, fmt.Errorf("queueing email change code: %w", err))
		uc.logEmailChange(ctx, userID, logger.OutcomeFailure, logger.ReasonDeliveryFailed)
		return nil
	}

	uc.logEmailChange(ctx, userID, logger.OutcomeSuccess, "")
	return nil
}

// ConfirmEmailChange moves the account to the address the code was sent to, marks it verified,
// and tells the old address. Every wrong code is ErrInvalidVerificationCode, under the same
// attempt ceiling as verification (see matchOTP).
//
// Sessions are kept. The change was authorised with the password and confirmed from the new
// inbox, and the old address is warned, which is where a user who did not make the change
// learns to reset their password and sign everyone out.
func (uc *authUseCase) ConfirmEmailChange(ctx context.Context, userID, code string) error {
	if uc.email.Notifier == nil {
		return errEmailFlowsUnavailable
	}

	user, err := uc.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}
	if user == nil || !user.IsActive {
		return ErrUnauthorized
	}

	token, reason, err := uc.matchOTP(ctx, userID, OneTimeTokenEmailChange, code)
	if err != nil {
		return err
	}
	if token == nil {
		uc.logEmailChanged(ctx, userID, logger.OutcomeFailure, reason)
		return ErrInvalidVerificationCode
	}

	err = uc.txManager.Do(ctx, func(txCtx context.Context) error {
		repo := uc.authRepo.WithTx(txCtx)

		if _, err := repo.ConsumeOneTimeToken(txCtx, token.TokenHash, OneTimeTokenEmailChange); err != nil {
			if errors.Is(err, ErrNotFound) {
				return ErrInvalidVerificationCode // a concurrent request with the same code won
			}
			return fmt.Errorf("consuming email change code: %w", err)
		}
		// The address was free when the code was sent; the constraint decides whether it still
		// is. Rolling back leaves the code usable, which is harmless while the address is held.
		if err := repo.UpdateUserEmail(txCtx, userID, token.NewEmail); err != nil {
			if errors.Is(err, ErrEmailAlreadyRegistered) {
				return err
			}
			return fmt.Errorf("updating email: %w", err)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrInvalidVerificationCode) {
			uc.logEmailChanged(ctx, userID, logger.OutcomeFailure, logger.ReasonInvalidToken)
		}
		if errors.Is(err, ErrEmailAlreadyRegistered) {
			uc.logEmailChanged(ctx, userID, logger.OutcomeFailure, logger.ReasonAddressTaken)
		}
		return err
	}

	uc.logEmailChanged(ctx, userID, logger.OutcomeSuccess, "")

	notice := EmailChangedNotice{OldEmail: user.Email, NewEmail: token.NewEmail, Name: user.Name}
	if err := uc.email.Notifier.SendEmailChanged(ctx, notice); err != nil {
		// The change has committed; failing the request now would only invite a retry against
		// a code that is spent.
		logger.Error(ctx, fmt.Errorf("queueing email changed notice: %w", err))
	}
	return nil
}

// SendAccountExistsNotice tells the owner of email that someone tried to register it again. It
// is how registration answers an address that is taken without saying so in the response.
func (uc *authUseCase) SendAccountExistsNotice(ctx context.Context, email string) error {
	if uc.email.Notifier == nil {
		return errEmailFlowsUnavailable
	}

	user, err := uc.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}
	if user == nil {
		return nil // deleted between the caller's check and this one; nobody to tell
	}

	return uc.email.Notifier.SendAccountExists(ctx, AccountExistsNotice{Email: user.Email, Name: user.Name})
}

func (uc *authUseCase) logEmailChange(ctx context.Context, userID, outcome, reason string) {
	logger.Security(ctx, logger.SecurityEvent{
		Action: logger.ActionEmailChangeRequested, Outcome: outcome, UserID: userID, Reason: reason,
	})
}

func (uc *authUseCase) logEmailChanged(ctx context.Context, userID, outcome, reason string) {
	logger.Security(ctx, logger.SecurityEvent{
		Action: logger.ActionEmailChanged, Outcome: outcome, UserID: userID, Reason: reason,
	})
}
