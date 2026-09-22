// Sign-in and sign-out: exchanging credentials for a session, and ending one.

package auth

import (
	"context"
	"errors"
	"fmt"
	"goilerplate/pkg/logger"
	"goilerplate/pkg/utils"
)

// Login authenticates a user and creates a session
func (uc *authUseCase) Login(ctx context.Context, credentials *LoginCredentials, deviceInfo *DeviceInfo) (*LoginResult, error) {
	// Validate user credentials
	user, err := uc.userValidator.ValidateUserForLogin(ctx, credentials.Email, credentials.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to validate user for login: %w", err)
	}

	// Generate session and tokens. The refresh token expires with the session, so the
	// session's absolute lifetime is decided here and handed to the signer. Signing touches
	// nothing outside this function, so it is done before the transaction opens.
	sessionID := utils.GenerateUUID()
	expiry := uc.sessionExpiry.For(credentials.RememberMe)
	tokenPair, err := uc.jwtService.GenerateTokenPair(
		user.ID,
		user.Name,
		user.Email,
		sessionID,
		deviceInfo.DeviceID,
		expiry,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	// Stamping the login and creating the session commit together: a failure between them
	// would otherwise leave a session the user was never told about, or a last_login_at with
	// no session behind it.
	session := uc.createUserSession(sessionID, user.ID, tokenPair.RefreshTokenID, deviceInfo, expiry)
	var createdSession *UserSession
	err = uc.txManager.Do(ctx, func(txCtx context.Context) error {
		repo := uc.authRepo.WithTx(txCtx)

		if err := repo.UpdateUserLoginInfo(txCtx, user.ID, true); err != nil {
			return fmt.Errorf("failed to update user login info: %w", err)
		}

		createdSession, err = repo.CreateSession(txCtx, session)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Logged after the commit, so the trail never claims a login that was rolled back. The IDs
	// are passed explicitly because this request authenticated no one until a moment ago:
	// the context still has no user or session on it.
	logger.Security(ctx, logger.SecurityEvent{
		Action:    logger.ActionLoginSucceeded,
		Outcome:   logger.OutcomeSuccess,
		UserID:    user.ID,
		SessionID: sessionID,
	})

	// Cache writes happen only after the commit, so a rolled-back login cannot leave the
	// cache describing a session that does not exist.
	if err := uc.permissionService.InvalidateUserPermissions(ctx, user.ID); err != nil {
		logger.Error(ctx, err)
	}

	menu, permissions, err := uc.buildMenuAndPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:       user,
		Menu:       menu,
		Permission: permissions,
		Tokens:     tokenPair,
		Session:    createdSession,
	}, nil
}

// buildMenuAndPermissions resolves the user's effective permissions and the menu tree they
// may see. Login and refresh both return it, so the client never has to reconcile the two.
func (uc *authUseCase) buildMenuAndPermissions(ctx context.Context, userID string) ([]Menu, []string, error) {
	// Resolved by the same service the per-request permission check uses. This function used to
	// fetch the roles, role permissions and overrides itself and merge them with its own copy of
	// the merge logic — so the list handed to the client at login and the list enforced on every
	// request afterwards were computed by two implementations that were free to drift apart.
	permissions, err := uc.permissionService.GetUserFinalPermissions(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to resolve user permissions: %w", err)
	}

	menus, err := uc.authRepo.GetParentMenus(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user menus: %w", err)
	}

	menuTree := uc.menuService.BuildMenuTree(ctx, menus)

	return uc.filterMenuTreeByPermissions(menuTree, permissions), permissions, nil
}

// filterMenuTreeByPermissions filters menu tree based on user permissions
func (uc *authUseCase) filterMenuTreeByPermissions(menuTree []Menu, userPermissions []string) []Menu {
	var filteredMenus []Menu

	// Create a map for faster permission lookup
	permissionMap := make(map[string]bool)
	for _, permission := range userPermissions {
		permissionMap[permission] = true
	}

	for _, menu := range menuTree {
		filteredMenu := uc.filterSingleMenu(menu, permissionMap)
		if filteredMenu != nil {
			filteredMenus = append(filteredMenus, *filteredMenu)
		}
	}

	return filteredMenus
}

// filterSingleMenu recursively filters a single menu and its children
func (uc *authUseCase) filterSingleMenu(menu Menu, permissionMap map[string]bool) *Menu {
	// Filter children recursively first
	var filteredChildren []Menu
	for _, child := range menu.Children {
		filteredChild := uc.filterSingleMenu(child, permissionMap)
		if filteredChild != nil {
			filteredChildren = append(filteredChildren, *filteredChild)
		}
	}

	// Check if this menu should be displayed
	var shouldDisplay bool

	if len(menu.Permissions) > 0 {
		// Menu has permissions, check if user has any of them
		hasAnyPermission := false
		for _, permissionID := range menu.Permissions {
			if permissionMap[permissionID] {
				hasAnyPermission = true
				break
			}
		}
		shouldDisplay = hasAnyPermission
	} else {
		// Menu doesn't have permissions (usually parent menu)
		// Display only if it has accessible children
		shouldDisplay = len(filteredChildren) > 0
	}

	if shouldDisplay {
		// Return the menu with filtered children
		filteredMenu := menu
		filteredMenu.Children = filteredChildren
		return &filteredMenu
	}

	return nil
}

// Logout revokes the caller's session, which is what invalidates both of its tokens: the
// access token stops passing the session check and the refresh token can no longer be
// exchanged. Other devices keep their own sessions.
// Note: Authentication is handled by middleware, userID and sessionID come from context.
func (uc *authUseCase) Logout(ctx context.Context, userID string, sessionID string) error {
	// Logging out twice is not an error: an already-revoked session is the state the caller
	// asked for. The cache is evicted either way, so a stale entry cannot outlive the call.
	if err := uc.authRepo.RevokeSession(ctx, userID, sessionID, RevokedReasonLogout); err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("revoking session: %w", err)
	}

	uc.sessionService.Evict(ctx, sessionID)

	logger.Security(ctx, logger.SecurityEvent{
		Action:    logger.ActionSessionRevoked,
		Outcome:   logger.OutcomeSuccess,
		UserID:    userID,
		SessionID: sessionID,
		Reason:    RevokedReasonLogout,
	})

	return nil
}

// LogoutAll invalidates all tokens for a user (logout from all devices)
// Note: Authentication is handled by middleware, userID comes from context
func (uc *authUseCase) LogoutAll(ctx context.Context, userID string) error {
	// Sessions are deactivated, not deleted, to keep an audit trail
	if err := uc.authRepo.DeactivateUserSessions(ctx, userID, RevokedReasonLogoutAll); err != nil {
		return fmt.Errorf("failed to deactivate user sessions: %w", err)
	}

	uc.sessionService.EvictUser(ctx, userID)

	logger.Security(ctx, logger.SecurityEvent{
		Action:  logger.ActionAllSessionsRevoked,
		Outcome: logger.OutcomeSuccess,
		UserID:  userID,
		Reason:  RevokedReasonLogoutAll,
	})

	return nil
}
