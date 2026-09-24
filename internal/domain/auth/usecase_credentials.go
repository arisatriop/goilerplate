// Credential management: creating an account, changing its password, and deactivating it.
// Each of these invalidates sessions, which is why they sit together.

package auth

import (
	"context"
	"fmt"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"
)

// Register creates a new user account.
//
// The plaintext arrives as its own argument rather than in entity.PasswordHash: a field named
// for a hash should never hold an unhashed value, however briefly.
func (uc *authUseCase) Register(ctx context.Context, entity *User, plaintextPassword string) error {
	existingUser, err := uc.authRepo.GetUserByEmail(ctx, entity.Email)
	if err != nil {
		return fmt.Errorf("failed to check if user exists: %w", err)
	}
	if existingUser != nil {
		return ErrEmailAlreadyRegistered
	}

	hashedPassword, err := utils.HashPassword(plaintextPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	entity.PasswordHash = hashedPassword
	entity.IsActive = true

	_, err = uc.authRepo.CreateUser(ctx, entity)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// ChangePassword replaces the caller's password and turns every other device out. The session
// making the change is kept, so the user is not logged out of the browser they are using.
//
// Knowing the current password is required: an access token alone is enough to be signed in,
// but not enough to take over an account someone else left signed in.
func (uc *authUseCase) ChangePassword(ctx context.Context, userID, sessionID, currentPassword, newPassword string) error {
	user, err := uc.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}
	if user == nil || !user.IsActive {
		return ErrUnauthorized
	}

	if err := utils.CheckPassword(currentPassword, user.PasswordHash); err != nil {
		return ErrInvalidCredentials
	}

	if err := uc.passwordPolicy.Validate(newPassword); err != nil {
		return err
	}

	hash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hashing new password: %w", err)
	}

	// The new password and the revocation commit together: a password changed without the
	// revocation would leave a thief signed in on another device with no sign anything happened.
	err = uc.txManager.Do(ctx, func(txCtx context.Context) error {
		repo := uc.authRepo.WithTx(txCtx)

		if err := repo.UpdateUserPassword(txCtx, userID, hash); err != nil {
			return fmt.Errorf("updating password: %w", err)
		}
		if err := repo.RevokeOtherUserSessions(txCtx, userID, sessionID, RevokedReasonPasswordChange); err != nil {
			return fmt.Errorf("revoking other sessions: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Evicting the whole user drops the surviving session too; it is re-read from the database
	// on the next request, which is correct and costs one lookup.
	uc.sessionService.EvictUser(ctx, userID)

	logger.Security(ctx, logger.SecurityEvent{
		Action:    logger.ActionPasswordChanged,
		Outcome:   logger.OutcomeSuccess,
		UserID:    userID,
		SessionID: sessionID,
	})

	return nil
}

// DeactivateUser disables an account and revokes its sessions in the same transaction. Doing
// only the first would leave the user with API access until their sessions expired, because the
// per-request check reads the session, not the account.
//
// This is the supported way to disable an account. Flipping users.is_active directly in the
// database bypasses it and leaves live sessions working.
func (uc *authUseCase) DeactivateUser(ctx context.Context, userID string) error {
	err := uc.txManager.Do(ctx, func(txCtx context.Context) error {
		repo := uc.authRepo.WithTx(txCtx)

		if err := repo.SetUserActive(txCtx, userID, false); err != nil {
			return fmt.Errorf("deactivating user: %w", err)
		}
		if err := repo.RevokeOtherUserSessions(txCtx, userID, "", RevokedReasonAdmin); err != nil {
			return fmt.Errorf("revoking sessions: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	uc.sessionService.EvictUser(ctx, userID)

	// The subject is the account being disabled, not whoever asked for it. Once an admin
	// endpoint calls this (it does not exist yet), the caller is identifiable through
	// request_id in the matching request log.
	logger.Security(ctx, logger.SecurityEvent{
		Action:  logger.ActionAccountDeactivated,
		Outcome: logger.OutcomeSuccess,
		UserID:  userID,
		Reason:  RevokedReasonAdmin,
	})

	return nil
}
