package repository

import (
	"context"
	"goilerplate/internal/domain/auth"
	"goilerplate/internal/infrastructure/model"
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

func (r *authRepository) CreateUser(ctx context.Context, user *auth.User) (*auth.User, error) {
	now := utils.TimeJakarta()
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
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.userModelToEntity(&data), nil
}

func (r *authRepository) IncrementFailedLoginAttempts(ctx context.Context, userID string) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Update("failed_login_attempts", gorm.Expr("failed_login_attempts + 1"))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *authRepository) LockUser(ctx context.Context, userID string, lockedUntil *time.Time) error {
	updates := map[string]interface{}{
		"locked_until": lockedUntil,
		"updated_at":   utils.TimeJakarta(),
	}

	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *authRepository) UpdateUserLoginInfo(ctx context.Context, userID string, resetFailedAttempts bool) error {
	now := utils.TimeJakarta()
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
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *authRepository) ResetExpiredLock(ctx context.Context, userID string) error {
	updates := map[string]interface{}{
		"failed_login_attempts": 0,
		"locked_until":          nil,
		"updated_at":            utils.TimeJakarta(),
	}

	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ? AND deleted_at IS NULL", userID).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// Session operations
func (r *authRepository) CreateSession(ctx context.Context, session *auth.UserSession) (*auth.UserSession, error) {
	sessionModel := &model.UserSession{
		ID:               session.ID,
		UserID:           session.UserID,
		RefreshTokenHash: session.RefreshTokenHash,
		DeviceName:       session.DeviceName,
		DeviceType:       session.DeviceType,
		DeviceID:         session.DeviceID,
		IPAddress:        session.IPAddress,
		UserAgent:        session.UserAgent,
		Location:         session.Location,
		IsActive:         session.IsActive,
		ExpiresAt:        session.ExpiresAt,
		LastUsedAt:       session.LastUsedAt,
	}

	if err := r.db.WithContext(ctx).Create(sessionModel).Error; err != nil {
		return nil, err
	}

	return session, nil
}

func (r *authRepository) GetSessionByID(ctx context.Context, sessionID string) (*auth.UserSession, error) {
	var sessionModel model.UserSession
	if err := r.db.WithContext(ctx).
		Where("id = ?", sessionID).
		First(&sessionModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &auth.UserSession{
		ID:               sessionModel.ID,
		UserID:           sessionModel.UserID,
		RefreshTokenHash: sessionModel.RefreshTokenHash,
		DeviceID:         sessionModel.DeviceID,
		DeviceName:       sessionModel.DeviceName,
		DeviceType:       sessionModel.DeviceType,
		IPAddress:        sessionModel.IPAddress,
		UserAgent:        sessionModel.UserAgent,
		Location:         sessionModel.Location,
		IsActive:         sessionModel.IsActive,
		ExpiresAt:        sessionModel.ExpiresAt,
		LastUsedAt:       sessionModel.LastUsedAt,
	}, nil
}

func (r *authRepository) DeleteUserSessions(ctx context.Context, userID string) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.UserSession{})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *authRepository) DeactivateUserSessions(ctx context.Context, userID string) error {
	result := r.db.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("user_id = ?", userID).
		Update("is_active", false)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// Token operations
func (r *authRepository) CreateToken(ctx context.Context, token *auth.UserToken) (*auth.UserToken, error) {
	tokenModel := &model.UserToken{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		TokenType: token.TokenType,
		ExpiresAt: token.ExpiresAt,
		UsedAt:    token.UsedAt,
		IPAddress: token.IPAddress,
		UserAgent: token.UserAgent,
	}

	if err := r.db.WithContext(ctx).Create(tokenModel).Error; err != nil {
		return nil, err
	}

	return token, nil
}

func (r *authRepository) GetTokenByHash(ctx context.Context, tokenHash string) (*auth.UserToken, error) {
	var tokenModel model.UserToken
	if err := r.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		First(&tokenModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.tokenModelToEntity(&tokenModel), nil
}

func (r *authRepository) DeleteTokenByHash(ctx context.Context, tokenHash string) error {
	result := r.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		Delete(&model.UserToken{})

	if result.Error != nil {
		return result.Error
	}

	// Check if any rows were actually deleted
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *authRepository) GetUserTokens(ctx context.Context, userID string) ([]auth.UserToken, error) {
	var tokenModels []model.UserToken
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&tokenModels).Error

	if err != nil {
		return nil, err
	}

	// Convert models to entities
	tokens := make([]auth.UserToken, len(tokenModels))
	for i, tokenModel := range tokenModels {
		tokens[i] = *r.tokenModelToEntity(&tokenModel)
	}

	return tokens, nil
}

func (r *authRepository) DeleteUserTokens(ctx context.Context, userID string) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&model.UserToken{})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *authRepository) DeleteTokensBySession(ctx context.Context, userID, sessionID string) error {
	// Find the session to get the refresh token hash
	var session model.UserSession
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", sessionID, userID).
		First(&session).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return gorm.ErrRecordNotFound // Return not found instead of nil
		}
		return err
	}

	// Delete the refresh token associated with this session
	result := r.db.WithContext(ctx).
		Where("token_hash = ?", session.RefreshTokenHash).
		Delete(&model.UserToken{})

	if result.Error != nil {
		return result.Error
	}

	// Check if any refresh token was actually deleted
	if result.RowsAffected == 0 {
		// No refresh token found, but session exists - this is unusual
		// Still deactivate the session for consistency
		err = r.db.WithContext(ctx).
			Model(&model.UserSession{}).
			Where("id = ?", sessionID).
			Update("is_active", false).Error

		if err != nil {
			return err
		}

		return gorm.ErrRecordNotFound // No tokens were deleted
	}

	// Deactivate the session
	err = r.db.WithContext(ctx).
		Model(&model.UserSession{}).
		Where("id = ?", sessionID).
		Update("is_active", false).Error

	return err
}

func (r *authRepository) GetUserByID(ctx context.Context, userID string) (*auth.User, error) {
	var data model.User

	if err := r.db.WithContext(ctx).
		Where("id = ? and deleted_at IS NULL", userID).
		First(&data).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return r.userModelToEntity(&data), nil
}

func (r *authRepository) MarkTokenAsUsed(ctx context.Context, token string) error {
	now := utils.TimeJakarta()
	result := r.db.WithContext(ctx).
		Model(&model.UserToken{}).
		Where("token_hash = ?", token).
		Update("used_at", now)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
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

// tokenModelToEntity converts model.UserToken to auth.UserToken entity
func (r *authRepository) tokenModelToEntity(m *model.UserToken) *auth.UserToken {
	if m == nil {
		return nil
	}

	return &auth.UserToken{
		ID:        m.ID,
		UserID:    m.UserID,
		TokenHash: m.TokenHash,
		TokenType: m.TokenType,
		ExpiresAt: m.ExpiresAt,
		UsedAt:    m.UsedAt,
		IPAddress: m.IPAddress,
		UserAgent: m.UserAgent,
	}
}

// GetRoleMenus retrieves only the parent menus from role_menus for a user
func (r *authRepository) GetRoleMenus(ctx context.Context, userID string) ([]auth.Menu, error) {
	var menus []model.Menu
	query := `
		SELECT DISTINCT m.*
		FROM menus m
		JOIN role_menus rm ON m.id = rm.menu_id
		JOIN user_roles ur ON rm.role_id = ur.role_id
		WHERE ur.user_id = ?
		AND m.is_active = true
		AND m.deleted_at IS NULL
		ORDER BY m.display_order ASC, m.name ASC
	`
	if err := r.db.WithContext(ctx).Raw(query, userID).Scan(&menus).Error; err != nil {
		return nil, err
	}

	// Convert models to entities
	entities := make([]auth.Menu, len(menus))
	for i, menu := range menus {
		entities[i] = r.menuModelToEntity(&menu)
	}

	return entities, nil
}

// GetUserMenus retrieves only the parent menus from user_menus for a user
func (r *authRepository) GetUserMenus(ctx context.Context, userID string) ([]auth.Menu, error) {
	var menus []model.Menu
	query := `
		SELECT DISTINCT m.*
		FROM menus m
		JOIN user_menus um ON m.id = um.menu_id
		WHERE um.user_id = ?
		AND m.is_active = true
		AND m.deleted_at IS NULL
		ORDER BY m.display_order ASC, m.name ASC
	`
	if err := r.db.WithContext(ctx).Raw(query, userID).Scan(&menus).Error; err != nil {
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
		Children:     []auth.Menu{}, // Will be populated by menu service
	}
}

// GetUserPermissionOverride checks if user has explicit permission override in user_permissions table
// Returns:
// - *bool: nil if no override exists, true if is_granted = true, false if is_granted = false
// - error: any database error
func (r *authRepository) GetUserPermissionOverride(ctx context.Context, userID string, permissionSlug string) (*bool, error) {
	var result struct {
		IsGranted bool
	}

	err := r.db.WithContext(ctx).
		Table("user_permissions up").
		Select("up.is_granted").
		Joins("JOIN permissions p ON up.permission_id = p.id").
		Where("up.user_id = ?", userID).
		Where("p.slug = ?", permissionSlug).
		Where("p.deleted_at IS NULL").
		First(&result).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No user-specific override
		}
		return nil, err
	}

	return &result.IsGranted, nil
}

// HasRolePermission checks if user has permission through their roles
// Path: users -> user_roles -> roles -> role_permissions -> permissions
func (r *authRepository) HasRolePermission(ctx context.Context, userID string, permissionSlug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("role_permissions rp").
		Joins("JOIN permissions p ON rp.permission_id = p.id").
		Joins("JOIN user_roles ur ON rp.role_id = ur.role_id").
		Where("ur.user_id = ?", userID).
		Where("p.slug = ?", permissionSlug).
		Where("p.deleted_at IS NULL").
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// HasMenuPermissionByMenuIDs checks if any of the given menu IDs have the specified permission
// This is a simple lookup without any tree traversal logic
func (r *authRepository) HasMenuPermissionByMenuIDs(ctx context.Context, menuIDs []string, permissionSlug string) (bool, error) {
	if len(menuIDs) == 0 {
		return false, nil
	}

	var count int64
	err := r.db.WithContext(ctx).
		Table("menu_permissions mp").
		Joins("JOIN permissions p ON mp.permission_id = p.id").
		Where("mp.menu_id IN ?", menuIDs).
		Where("p.slug = ?", permissionSlug).
		Where("p.deleted_at IS NULL").
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
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

// GetRolePermissionSlugs gets all permission slugs from role_permissions via user's roles
func (r *authRepository) GetRolePermissionSlugs(ctx context.Context, userID string) ([]string, error) {
	var slugs []string
	err := r.db.WithContext(ctx).
		Table("role_permissions rp").
		Select("DISTINCT p.slug").
		Joins("JOIN permissions p ON rp.permission_id = p.id").
		Joins("JOIN user_roles ur ON rp.role_id = ur.role_id").
		Where("ur.user_id = ?", userID).
		Where("p.deleted_at IS NULL").
		Pluck("p.slug", &slugs).Error

	if err != nil {
		return nil, err
	}

	return slugs, nil
}

// GetMenuPermissionSlugs gets all permission slugs from menu_permissions for given menu IDs
func (r *authRepository) GetMenuPermissionSlugs(ctx context.Context, menuIDs []string) ([]string, error) {
	if len(menuIDs) == 0 {
		return []string{}, nil
	}

	var slugs []string
	err := r.db.WithContext(ctx).
		Table("menu_permissions mp").
		Select("DISTINCT p.slug").
		Joins("JOIN permissions p ON mp.permission_id = p.id").
		Where("mp.menu_id IN ?", menuIDs).
		Where("p.deleted_at IS NULL").
		Pluck("p.slug", &slugs).Error

	if err != nil {
		return nil, err
	}

	return slugs, nil
}

// GetAllUserPermissionOverrides gets all user_permissions (both grants and revocations)
// Returns map[permissionSlug]isGranted
func (r *authRepository) GetAllUserPermissionOverrides(ctx context.Context, userID string) (map[string]bool, error) {
	type UserPermission struct {
		Slug      string
		IsGranted bool
	}

	var results []UserPermission
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
