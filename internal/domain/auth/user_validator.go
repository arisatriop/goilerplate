package auth

import (
	"context"
	"fmt"
	"goilerplate/pkg/constants"
	"goilerplate/pkg/utils"
	"net/http"
	"time"
)

// UserValidator handles user validation logic
type UserValidator struct {
	authRepo Repository
}

// NewUserValidator creates a new user validator
func NewUserValidator(authRepo Repository) *UserValidator {
	return &UserValidator{
		authRepo: authRepo,
	}
}

// ValidateUserForLogin validates user for login operation
func (uv *UserValidator) ValidateUserForLogin(ctx context.Context, email, password string) (*User, error) {
	user, err := uv.authRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	if user == nil {
		return nil, utils.Error(http.StatusBadRequest, constants.MsgInvalidCredential)
	}

	if !user.IsActive {
		return nil, utils.Error(http.StatusForbidden, constants.MsgAccountDisabled)
	}

	// Reset failed attempts if the lock has expired
	if user.HasExpiredLock() {
		if err := uv.authRepo.ResetExpiredLock(ctx, user.ID); err != nil {
			return nil, fmt.Errorf("failed to reset expired lock: %w", err)
		}
		// Update user object to reflect the reset
		user.FailedLoginAttempts = 0
		user.LockedUntil = nil
	}

	if user.IsLocked() {
		return nil, utils.Error(http.StatusForbidden, constants.MsgAccountLocked)
	}

	if err := uv.verifyPassword(ctx, password, user); err != nil {
		return nil, fmt.Errorf("failed to veify password: %w", err)
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
		return nil, utils.Error(http.StatusNotFound, constants.MsgResourceNotFound)
	}

	if !user.IsActive {
		return nil, utils.Error(http.StatusForbidden, constants.MsgAccountDisabled)
	}

	return user, nil
}

// verifyPassword verifies the user password and handles failed attempts
func (uv *UserValidator) verifyPassword(ctx context.Context, password string, user *User) error {
	err := utils.CheckPassword(password, user.PasswordHash)
	if err != nil {
		// Increment failed login attempts
		if incrementErr := uv.authRepo.IncrementFailedLoginAttempts(ctx, user.ID); incrementErr != nil {
			return fmt.Errorf("failed to increment failed login attempts: %w", incrementErr)
		}

		// Lock account if too many failed attempts
		if user.ShouldLockAccount(MaxFailedLoginAttempts) {
			lockUntil := utils.Now().Add(time.Duration(AccountLockDuration) * time.Minute)
			if lockErr := uv.authRepo.LockUser(ctx, user.ID, &lockUntil); lockErr != nil {
				return fmt.Errorf("failed to lock user account: %w", lockErr)
			}
		}

		return utils.Error(http.StatusBadRequest, constants.MsgInvalidCredential)
	}
	return nil
}
