package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService handles caching operations for auth data
type CacheService struct {
	redis   *redis.Client
	enabled bool
}

// NewCacheService creates a new cache service
func NewCacheService(redis *redis.Client) *CacheService {
	return &CacheService{
		redis:   redis,
		enabled: redis != nil,
	}
}

// IsEnabled returns whether Redis caching is enabled
func (s *CacheService) IsEnabled() bool {
	return s.enabled
}

// CacheSession stores session data in Redis
func (s *CacheService) CacheSession(ctx context.Context, session *UserSession) error {
	key := fmt.Sprintf("session:%s", session.ID)
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	// Set TTL based on session expiration
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}

	err = s.redis.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache session: %w", err)
	}

	return nil
}

// CacheToken stores token data in Redis
func (s *CacheService) CacheToken(ctx context.Context, token *UserToken) error {
	if !s.enabled {
		return nil // Skip if Redis is disabled
	}

	key := fmt.Sprintf("token:%s", token.TokenHash)
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Set TTL based on token expiration
	ttl := time.Until(token.ExpiresAt)
	if ttl <= 0 {
		ttl = 24 * time.Hour // Default to 24 hours if expired
	}

	err = s.redis.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache token: %w", err)
	}

	return nil
}

// GetSession retrieves session from Redis
func (s *CacheService) GetSession(ctx context.Context, sessionID string) (*UserSession, error) {
	if !s.enabled {
		return nil, nil // Return nil if Redis is disabled
	}

	key := fmt.Sprintf("session:%s", sessionID)
	data, err := s.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Session not found in cache
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session from cache: %w", err)
	}

	var session UserSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// GetToken retrieves token from Redis
func (s *CacheService) GetToken(ctx context.Context, tokenHash string) (*UserToken, error) {
	if !s.enabled {
		return nil, nil // Return nil if Redis is disabled
	}

	key := fmt.Sprintf("token:%s", tokenHash)
	data, err := s.redis.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Token not found in cache
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get token from cache: %w", err)
	}

	var token UserToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}

// DeleteSession removes session from Redis
func (s *CacheService) DeleteSession(ctx context.Context, sessionID string) error {
	if !s.enabled {
		return nil // Skip if Redis is disabled
	}

	key := fmt.Sprintf("session:%s", sessionID)
	err := s.redis.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete session from cache: %w", err)
	}

	return nil
}

// DeleteToken removes token from Redis
func (s *CacheService) DeleteToken(ctx context.Context, tokenHash string) error {
	if !s.enabled {
		return nil // Skip if Redis is disabled
	}

	key := fmt.Sprintf("token:%s", tokenHash)
	err := s.redis.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete token from cache: %w", err)
	}

	return nil
}

// DeleteUserSessions removes all sessions for a user
func (s *CacheService) DeleteUserSessions(ctx context.Context, userID string) error {
	if !s.enabled {
		return nil // Skip if Redis is disabled
	}

	pattern := "session:*"
	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()
		// Get session data to check user ID
		data, err := s.redis.Get(ctx, key).Bytes()
		if err != nil {
			continue // Skip on error
		}

		var session UserSession
		if err := json.Unmarshal(data, &session); err != nil {
			continue // Skip on error
		}

		// Delete if it belongs to the user
		if session.UserID == userID {
			s.redis.Del(ctx, key)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan sessions: %w", err)
	}

	return nil
}

// DeleteUserTokens removes all tokens for a user
func (s *CacheService) DeleteUserTokens(ctx context.Context, userID string) error {
	if !s.enabled {
		return nil // Skip if Redis is disabled
	}

	pattern := "token:*"
	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()
		// Get token data to check user ID
		data, err := s.redis.Get(ctx, key).Bytes()
		if err != nil {
			continue // Skip on error
		}

		var token UserToken
		if err := json.Unmarshal(data, &token); err != nil {
			continue // Skip on error
		}

		// Delete if it belongs to the user
		if token.UserID == userID {
			s.redis.Del(ctx, key)
		}
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan tokens: %w", err)
	}

	return nil
}

// Permission caching methods

// CacheUserPermissions caches a user's final permission list
// Key format: "permission:{userID}"
func (s *CacheService) CacheUserPermissions(ctx context.Context, userID string, permissions map[string]struct{}, ttl time.Duration) error {
	if !s.enabled {
		return nil // Skip if Redis is disabled
	}

	key := fmt.Sprintf("permission:%s", userID)

	data, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	err = s.redis.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache permission: %w", err)
	}

	return nil
}

// GetCachedUserPermissions retrieves cached user permissions
// Returns (permissions []string, found bool, error)
func (s *CacheService) GetCachedUserPermissions(ctx context.Context, userID string) ([]string, bool, error) {
	if !s.enabled {
		return nil, false, nil // Skip if Redis is disabled
	}

	key := fmt.Sprintf("permission:%s", userID)
	result, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false, nil // Not found in cache
	}
	if err != nil {
		return nil, false, fmt.Errorf("failed to get cached permissions: %w", err)
	}

	var permissionMap map[string]struct{}
	err = json.Unmarshal([]byte(result), &permissionMap)
	if err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	// Convert map back to slice
	permissions := make([]string, 0, len(permissionMap))
	for permission := range permissionMap {
		permissions = append(permissions, permission)
	}

	return permissions, true, nil
}

// CacheUserPermission caches a user's permission check result
// Key format: "permission:{userID}"
func (s *CacheService) CacheUserPermission(ctx context.Context, userID string, permissions map[string]struct{}, ttl time.Duration) error {
	key := fmt.Sprintf("permission:%s", userID)

	data, err := json.Marshal(permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	err = s.redis.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache permission: %w", err)
	}

	return nil
}

// GetCachedUserPermission retrieves cached permission check result
// Returns (hasPermission bool, error)
func (s *CacheService) GetCachedUserPermission(ctx context.Context, userID string, permissionSlug string) (bool, error) {
	if !s.enabled {
		return false, fmt.Errorf("redis is disabled")
	}

	key := fmt.Sprintf("permission:%s", userID)
	result, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, fmt.Errorf("permission cache not found")
	}
	if err != nil {
		return false, fmt.Errorf("failed to get cached permission: %w", err)
	}

	var permissions map[string]struct{}
	err = json.Unmarshal([]byte(result), &permissions)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	if _, ok := permissions[permissionSlug]; ok {
		return true, nil
	}

	return false, nil
}

// InvalidateAllPermissions removes all cached permissions (for all users)
// This should be called when roles, permissions, or menus are modified globally
func (s *CacheService) InvalidateAllPermissions(ctx context.Context) error {
	if !s.enabled {
		return nil // Skip if Redis is disabled
	}

	pattern := "permission:*"
	iter := s.redis.Scan(ctx, 0, pattern, 0).Iterator()

	for iter.Next(ctx) {
		key := iter.Val()
		s.redis.Del(ctx, key)
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan permissions: %w", err)
	}

	return nil
}

// InvalidateUserPermissions removes cached permissions for a specific user
// This should be called when user's roles or permissions are modified
func (s *CacheService) InvalidateUserPermissions(ctx context.Context, userID string) error {
	if !s.enabled {
		return nil // Skip if Redis is disabled
	}

	key := fmt.Sprintf("permission:%s", userID)
	err := s.redis.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to invalidate user permissions: %w", err)
	}

	return nil
}

// Token Blacklist methods

// AddTokenToBlacklist adds a token hash to the blacklist with TTL
// The TTL should be set to the remaining time until token expiration
// This ensures the blacklist entry is automatically cleaned up after the token would have expired anyway
func (s *CacheService) AddTokenToBlacklist(ctx context.Context, tokenHash string, ttl time.Duration) error {
	if !s.enabled {
		return nil // Skip if Redis is disabled (rely on DB deletion)
	}

	key := fmt.Sprintf("blacklist:%s", tokenHash)
	// Value doesn't matter, we just check for key existence
	// Set TTL to token's remaining lifetime
	err := s.redis.Set(ctx, key, "1", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}

// IsTokenBlacklisted checks if a token hash exists in the blacklist
func (s *CacheService) IsTokenBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	if !s.enabled {
		return false, nil // If Redis is disabled, rely on DB lookup
	}

	key := fmt.Sprintf("blacklist:%s", tokenHash)
	exists, err := s.redis.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return exists > 0, nil
}

// AddMultipleTokensToBlacklist adds multiple token hashes to blacklist
// Useful for logout-all scenarios
func (s *CacheService) AddMultipleTokensToBlacklist(ctx context.Context, tokenHashes []string, ttl time.Duration) error {
	if !s.enabled {
		return nil // Skip if Redis is disabled
	}

	if len(tokenHashes) == 0 {
		return nil
	}

	// Use pipeline for better performance
	pipe := s.redis.Pipeline()
	for _, tokenHash := range tokenHashes {
		key := fmt.Sprintf("blacklist:%s", tokenHash)
		pipe.Set(ctx, key, "1", ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to add multiple tokens to blacklist: %w", err)
	}

	return nil
}
