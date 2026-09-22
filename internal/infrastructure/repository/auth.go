package repository

import (
	"context"
	"errors"
	"goilerplate/internal/domain/auth"
	"goilerplate/internal/infrastructure/model"
	"goilerplate/internal/infrastructure/transaction"
	"goilerplate/pkg/utils"
	"time"

	"gorm.io/gorm"
)

type authRepository struct {
	db *gorm.DB
}

func NewAuth(db *gorm.DB) auth.Repository {
	return &authRepository{
		db: db,
	}
}

func (r *authRepository) WithTx(ctx context.Context) auth.Repository {
	if tx := transaction.GetTxFromContext(ctx); tx != nil {
		return NewAuth(tx)
	}
	return r
}

func (r *authRepository) CreateUser(ctx context.Context, user *auth.User) (*auth.User, error) {
	now := utils.Now()
	id := utils.GenerateUUID()
	model := &model.User{
		ID:                  id,
		Name:                user.Name,
		Email:               user.Email,
		Avatar:              user.Avatar,
		PasswordHash:        user.PasswordHash,
		IsActive:            user.IsActive,
		EmailVerified:       user.EmailVerified,
		EmailVerifiedAt:     user.EmailVerifiedAt,
		PasswordChangedAt:   now, // Set to current time on creation
		LastLoginAt:         user.LastLoginAt,
		FailedLoginAttempts: user.FailedLoginAttempts,
		LockedUntil:         user.LockedUntil,
		CreatedAt:           now,
		CreatedBy:           id,
		UpdatedAt:           now,
		UpdatedBy:           id,
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}

	return r.userModelToEntity(model), nil
}

func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	var data model.User

	if err := r.db.WithContext(ctx).
		Where("email = ? and deleted_at IS NULL", email).
		First(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.userModelToEntity(&data), nil
}

// RegisterFailedLogin counts one failed attempt and, on the attempt that reaches maxAttempts,
// locks the account — in a single statement, so simultaneous guesses cannot each read the same
// stale counter and skip past the threshold. The CASE compares the post-increment value, so the
// lock lands exactly on attempt N rather than N+1. It returns whether the account is now locked.
func (r *authRepository) RegisterFailedLogin(ctx context.Context, userID string, maxAttempts int, lockUntil time.Time) (bool, error) {
	var row struct {
		FailedLoginAttempts int
		LockedUntil         *time.Time
	}

	err := r.db.WithContext(ctx).Raw(`
		UPDATE users
		   SET failed_login_attempts = failed_login_attempts + 1,
		       locked_until = CASE WHEN failed_login_attempts + 1 >= ? THEN ? ELSE locked_until END,
		       updated_at = ?
		 WHERE id = ? AND deleted_at IS NULL
		RETURNING failed_login_attempts, locked_until`,
		maxAttempts, lockUntil, utils.Now(), userID,
	).Scan(&row).Error
	if err != nil {
		return false, err
	}
	if row.FailedLoginAttempts == 0 {
		return false, auth.ErrNotFound
	}

	return row.LockedUntil != nil && row.LockedUntil.After(utils.Now()), nil
}

func (r *authRepository) UpdateUserLoginInfo(ctx context.Context, userID string, resetFailedAttempts bool) error {
	now := utils.Now()
	updates := map[string]interface{}{
		"last_login_at": now,
		"updated_at":    now,
	}

	if resetFailedAttempts {
		updates["failed_login_attempts"] = 0
		updates["locked_until"] = nil
	}

	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return auth.ErrNotFound
	}

	return nil
}

func (r *authRepository) ResetExpiredLock(ctx context.Context, userID string) error {
	updates := map[string]interface{}{
		"failed_login_attempts": 0,
		"locked_until":          nil,
		"updated_at":            utils.Now(),
	}

	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return auth.ErrNotFound
	}

	return nil
}

// Session operations
func (r *authRepository) CreateSession(ctx context.Context, session *auth.UserSession) (*auth.UserSession, error) {
	sessionModel := &model.UserSession{
		ID:                 session.ID,
		UserID:             session.UserID,
		RefreshJTI:         session.RefreshJTI,
		PreviousRefreshJTI: nullableString(session.PreviousRefreshJTI),
		RotatedAt:          session.RotatedAt,
		DeviceName:         session.DeviceName,
		DeviceType:         session.DeviceType,
		DeviceID:           session.DeviceID,
		IPAddress:          nullableString(session.IPAddress),
		UserAgent:          session.UserAgent,
		IsActive:           session.IsActive,
		ExpiresAt:          session.ExpiresAt,
		LastUsedAt:         session.LastUsedAt,
		RevokedAt:          session.RevokedAt,
		RevokedReason:      nullableString(session.RevokedReason),
		CreatedAt:          utils.Now(),
	}

	if err := r.db.WithContext(ctx).Create(sessionModel).Error; err != nil {
		return nil, err
	}

	session.CreatedAt = sessionModel.CreatedAt
	return session, nil
}

func (r *authRepository) GetSessionByID(ctx context.Context, sessionID string) (*auth.UserSession, error) {
	var sessionModel model.UserSession
	if err := r.db.WithContext(ctx).
		Where("id = ?", sessionID).
		First(&sessionModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return userSessionModelToEntity(&sessionModel), nil
}

func (r *authRepository) DeactivateUserSessions(ctx context.Context, userID, reason string) error {
	result := r.db.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("user_id = ? AND is_active", userID).
		Updates(map[string]any{
			"is_active":      false,
			"revoked_at":     utils.Now(),
			"revoked_reason": reason,
		})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// RotateRefreshJTI claims the refresh token named by currentJTI and replaces it with newJTI,
// in one conditional UPDATE. Only the request whose jti still matches wins, so two concurrent
// refreshes cannot both rotate. It returns auth.ErrNotFound when nothing matched, which means
// either the token was already rotated away or the session is no longer usable — the caller
// re-reads the session to tell those apart.
func (r *authRepository) RotateRefreshJTI(ctx context.Context, sessionID, currentJTI, newJTI string) error {
	now := utils.Now()

	result := r.db.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("id = ? AND refresh_jti = ? AND is_active AND expires_at > ?", sessionID, currentJTI, now).
		Updates(map[string]any{
			"previous_refresh_jti": currentJTI,
			"refresh_jti":          newJTI,
			"rotated_at":           now,
			"last_used_at":         now,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return auth.ErrNotFound
	}

	return nil
}

// RevokeOtherUserSessions revokes every active session of a user except keepSessionID. It is
// what a password change uses: the device making the change stays signed in, every other one is
// turned out. Pass an empty keepSessionID to revoke all of them.
func (r *authRepository) RevokeOtherUserSessions(ctx context.Context, userID, keepSessionID, reason string) error {
	query := r.db.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("user_id = ? AND is_active", userID)

	if keepSessionID != "" {
		query = query.Where("id <> ?", keepSessionID)
	}

	return query.Updates(map[string]any{
		"is_active":      false,
		"revoked_at":     utils.Now(),
		"revoked_reason": reason,
	}).Error
}

// UpdateUserPassword stores a new password hash and stamps password_changed_at, which is the
// record of when every other session was turned out.
func (r *authRepository) UpdateUserPassword(ctx context.Context, userID, passwordHash string) error {
	now := utils.Now()

	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Updates(map[string]any{
			"password_hash":         passwordHash,
			"password_changed_at":   now,
			"failed_login_attempts": 0,
			"locked_until":          nil,
			"updated_at":            now,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return auth.ErrNotFound
	}

	return nil
}

// SetUserActive flips the account's active flag. Revoking the user's sessions is the caller's
// job, and must happen in the same transaction: a deactivated account with live sessions would
// keep API access until they expired.
func (r *authRepository) SetUserActive(ctx context.Context, userID string, active bool) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Updates(map[string]any{"is_active": active, "updated_at": utils.Now()})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return auth.ErrNotFound
	}

	return nil
}

// RevokeSession deactivates one session in a single conditional UPDATE, so concurrent
// logouts cannot both report success. It returns auth.ErrNotFound when the session does not
// belong to the user or was already revoked.
func (r *authRepository) RevokeSession(ctx context.Context, userID, sessionID, reason string) error {
	now := utils.Now()

	result := r.db.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("id = ? AND user_id = ? AND is_active", sessionID, userID).
		Updates(map[string]any{
			"is_active":      false,
			"revoked_at":     now,
			"revoked_reason": reason,
		})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return auth.ErrNotFound
	}

	return nil
}

func (r *authRepository) GetUserByID(ctx context.Context, userID string) (*auth.User, error) {
	var data model.User

	if err := r.db.WithContext(ctx).
		Where("id = ? and deleted_at IS NULL", userID).
		First(&data).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return r.userModelToEntity(&data), nil
}

// GetParentMenus retrieves all parent menus (menus without parent_id)
func (r *authRepository) GetParentMenus(ctx context.Context) ([]auth.Menu, error) {
	var menus []model.Menu
	query := `
		SELECT m.*
		FROM menus m
		WHERE m.is_active = true
		AND m.deleted_at IS NULL
		AND m.parent_id IS NULL
		ORDER BY m.display_order ASC, m.name ASC
	`
	if err := r.db.WithContext(ctx).Raw(query).Scan(&menus).Error; err != nil {
		return nil, err
	}

	// Convert models to entities
	entities := make([]auth.Menu, len(menus))
	for i, menu := range menus {
		entities[i] = r.menuModelToEntity(&menu)
	}

	return entities, nil
}

// GetMenusByParentIDs retrieves all child menus for the given parent IDs
func (r *authRepository) GetMenusByParentIDs(ctx context.Context, parentIDs []string) ([]auth.Menu, error) {
	if len(parentIDs) == 0 {
		return []auth.Menu{}, nil
	}

	var menus []model.Menu
	query := `
		SELECT m.*
		FROM menus m
		WHERE m.parent_id IN (?)
		AND m.is_active = true
		AND m.deleted_at IS NULL
		ORDER BY m.display_order ASC, m.name ASC
	`
	if err := r.db.WithContext(ctx).Raw(query, parentIDs).Scan(&menus).Error; err != nil {
		return nil, err
	}

	// Convert models to entities
	entities := make([]auth.Menu, len(menus))
	for i, menu := range menus {
		entities[i] = r.menuModelToEntity(&menu)
	}

	return entities, nil
}

// menuModelToEntity converts model.Menu to auth.Menu entity
func (r *authRepository) menuModelToEntity(m *model.Menu) auth.Menu {
	return auth.Menu{
		ID:           m.ID,
		ParentID:     m.ParentID,
		Name:         m.Name,
		Slug:         m.Slug,
		Icon:         m.Icon,
		Route:        m.Route,
		DisplayOrder: m.DisplayOrder,
		IsActive:     m.IsActive,
		Permissions:  []string{},    // Will be populated by menu service
		Children:     []auth.Menu{}, // Will be populated by menu service
	}
}

// GetUserRolesByUserID gets all role IDs for a user
func (r *authRepository) GetUserRolesByUserID(ctx context.Context, userID string) ([]string, error) {
	var roleIDs []string
	err := r.db.WithContext(ctx).
		Table("user_roles").
		Select("role_id").
		Where("user_id = ?", userID).
		Pluck("role_id", &roleIDs).Error

	if err != nil {
		return nil, err
	}

	return roleIDs, nil
}

// GetRolePermissionsByRoleIDs gets all permission slugs for given role IDs
func (r *authRepository) GetRolePermissionsByRoleIDs(ctx context.Context, roleIDs []string) ([]string, error) {
	if len(roleIDs) == 0 {
		return []string{}, nil
	}

	var permissionSlugs []string
	err := r.db.WithContext(ctx).
		Table("role_permissions rp").
		Select("DISTINCT p.slug").
		Joins("JOIN permissions p ON rp.permission_id = p.id").
		Where("rp.role_id IN ?", roleIDs).
		Where("p.deleted_at IS NULL").
		Pluck("p.slug", &permissionSlugs).Error

	if err != nil {
		return nil, err
	}

	return permissionSlugs, nil
}

// GetUserPermissionSlugs gets all permission slugs from user_permissions where is_granted = true
func (r *authRepository) GetUserPermissionSlugs(ctx context.Context, userID string) ([]string, error) {
	var slugs []string
	err := r.db.WithContext(ctx).
		Table("user_permissions up").
		Select("p.slug").
		Joins("JOIN permissions p ON up.permission_id = p.id").
		Where("up.user_id = ?", userID).
		Where("up.is_granted = ?", true).
		Where("p.deleted_at IS NULL").
		Pluck("p.slug", &slugs).Error

	if err != nil {
		return nil, err
	}

	return slugs, nil
}

// GetUserPermissionOverrides gets all user_permissions (both grants and revocations)
// Returns map[permissionSlug]isGranted
func (r *authRepository) GetUserPermissionOverrides(ctx context.Context, userID string) (map[string]bool, error) {
	type UserPermissionOverride struct {
		Slug      string
		IsGranted bool
	}

	var results []UserPermissionOverride
	err := r.db.WithContext(ctx).
		Table("user_permissions up").
		Select("p.slug, up.is_granted").
		Joins("JOIN permissions p ON up.permission_id = p.id").
		Where("up.user_id = ?", userID).
		Where("p.deleted_at IS NULL").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Convert to map
	overrides := make(map[string]bool)
	for _, result := range results {
		overrides[result.Slug] = result.IsGranted
	}

	return overrides, nil
}

// GetMenuPermissionsByMenuID gets all permission slugs for a single menu ID
func (r *authRepository) GetMenuPermissionsByMenuID(ctx context.Context, menuID string) ([]string, error) {
	var permissionSlugs []string
	err := r.db.WithContext(ctx).
		Table("menu_permissions mp").
		Select("p.slug").
		Joins("JOIN permissions p ON mp.permission_id = p.id").
		Where("mp.menu_id = ?", menuID).
		Where("p.deleted_at IS NULL").
		Pluck("p.slug", &permissionSlugs).Error

	if err != nil {
		return nil, err
	}

	return permissionSlugs, nil
}

// usermodelToEntity converts model.User to auth.User entity
func (r *authRepository) userModelToEntity(m *model.User) *auth.User {
	if m == nil {
		return nil
	}

	return &auth.User{
		ID:                  m.ID,
		Name:                m.Name,
		Email:               m.Email,
		Avatar:              m.Avatar,
		PasswordHash:        m.PasswordHash,
		IsActive:            m.IsActive,
		EmailVerified:       m.EmailVerified,
		EmailVerifiedAt:     m.EmailVerifiedAt,
		PasswordChangedAt:   m.PasswordChangedAt,
		LastLoginAt:         m.LastLoginAt,
		FailedLoginAttempts: m.FailedLoginAttempts,
		LockedUntil:         m.LockedUntil,
	}
}

// One-time token operations

func (r *authRepository) CreateOneTimeToken(ctx context.Context, token *auth.OneTimeToken) error {
	tokenModel := &model.OneTimeToken{
		ID:        utils.GenerateUUID(),
		UserID:    token.UserID,
		TokenType: token.TokenType,
		TokenHash: token.TokenHash,
		Attempts:  token.Attempts,
		ExpiresAt: token.ExpiresAt,
		UsedAt:    token.UsedAt,
		IPAddress: nullableString(token.IPAddress),
		UserAgent: token.UserAgent,
		CreatedAt: utils.Now(),
	}

	if err := r.db.WithContext(ctx).Create(tokenModel).Error; err != nil {
		return err
	}

	token.ID = tokenModel.ID
	token.CreatedAt = tokenModel.CreatedAt
	return nil
}

// GetLatestActiveOneTimeToken returns the newest unused, unexpired token of that type,
// or nil when there is none. OTP verification looks the token up this way so wrong guesses
// can be counted; a lookup by hash alone cannot.
func (r *authRepository) GetLatestActiveOneTimeToken(ctx context.Context, userID, tokenType string) (*auth.OneTimeToken, error) {
	var tokenModel model.OneTimeToken

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND token_type = ? AND used_at IS NULL AND expires_at > ?", userID, tokenType, utils.Now()).
		Order("created_at DESC").
		First(&tokenModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return oneTimeTokenModelToEntity(&tokenModel), nil
}

// ConsumeOneTimeToken marks a token as used in one statement, so two concurrent requests
// with the same token yield exactly one success. It returns auth.ErrNotFound when the token
// does not exist, was already used, or has expired.
func (r *authRepository) ConsumeOneTimeToken(ctx context.Context, tokenHash, tokenType string) error {
	now := utils.Now()

	result := r.db.WithContext(ctx).
		Model(&model.OneTimeToken{}).
		Where("token_hash = ? AND token_type = ? AND used_at IS NULL AND expires_at > ?", tokenHash, tokenType, now).
		Update("used_at", now)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return auth.ErrNotFound
	}

	return nil
}

// IncrementOneTimeTokenAttempts counts a failed verification attempt and returns the new total.
func (r *authRepository) IncrementOneTimeTokenAttempts(ctx context.Context, tokenID string) (int, error) {
	var attempts int

	err := r.db.WithContext(ctx).
		Raw(`UPDATE one_time_tokens SET attempts = attempts + 1
			WHERE id = ? AND used_at IS NULL
			RETURNING attempts`, tokenID).
		Scan(&attempts).Error
	if err != nil {
		return 0, err
	}
	if attempts == 0 {
		return 0, auth.ErrNotFound
	}

	return attempts, nil
}

func oneTimeTokenModelToEntity(m *model.OneTimeToken) *auth.OneTimeToken {
	token := &auth.OneTimeToken{
		ID:        m.ID,
		UserID:    m.UserID,
		TokenType: m.TokenType,
		TokenHash: m.TokenHash,
		Attempts:  m.Attempts,
		ExpiresAt: m.ExpiresAt,
		UsedAt:    m.UsedAt,
		UserAgent: m.UserAgent,
		CreatedAt: m.CreatedAt,
	}
	if m.IPAddress != nil {
		token.IPAddress = *m.IPAddress
	}
	return token
}

func userSessionModelToEntity(m *model.UserSession) *auth.UserSession {
	session := &auth.UserSession{
		ID:         m.ID,
		UserID:     m.UserID,
		RefreshJTI: m.RefreshJTI,
		RotatedAt:  m.RotatedAt,
		DeviceID:   m.DeviceID,
		DeviceName: m.DeviceName,
		DeviceType: m.DeviceType,
		UserAgent:  m.UserAgent,
		IsActive:   m.IsActive,
		ExpiresAt:  m.ExpiresAt,
		LastUsedAt: m.LastUsedAt,
		RevokedAt:  m.RevokedAt,
		CreatedAt:  m.CreatedAt,
	}
	if m.RevokedReason != nil {
		session.RevokedReason = *m.RevokedReason
	}
	if m.PreviousRefreshJTI != nil {
		session.PreviousRefreshJTI = *m.PreviousRefreshJTI
	}
	if m.IPAddress != nil {
		session.IPAddress = *m.IPAddress
	}
	return session
}

// nullableString stores an empty string as NULL, which typed columns such as inet require.
func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
