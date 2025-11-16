package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// TokenManager manages OAuth2 tokens with automatic refresh and caching
type TokenManager struct {
	config       *oauth2.Config
	token        *oauth2.Token
	tokenMu      sync.RWMutex
	cacheFile    string
	refreshToken string
}

// NewTokenManager creates a new token manager with caching support
func NewTokenManager(clientID, clientSecret, refreshToken, cacheDir string) (*TokenManager, error) {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		Scopes:       []string{"https://www.googleapis.com/auth/drive.file"},
	}

	tm := &TokenManager{
		config:       config,
		refreshToken: refreshToken,
		cacheFile:    filepath.Join(cacheDir, "gdrive-oauth-token.cache"),
	}

	// Try to load cached token
	if err := tm.loadCachedToken(); err != nil {
		// If no cached token, create initial token with refresh token
		tm.token = &oauth2.Token{
			RefreshToken: refreshToken,
		}
	}

	return tm, nil
}

// GetToken returns a valid access token, refreshing if necessary
// This method is thread-safe and handles automatic refresh
func (tm *TokenManager) GetToken(ctx context.Context) (*oauth2.Token, error) {
	tm.tokenMu.Lock()
	defer tm.tokenMu.Unlock()

	// Check if token is still valid (with 5 minute buffer)
	if tm.token != nil && tm.token.Valid() && time.Until(tm.token.Expiry) > 5*time.Minute {
		return tm.token, nil
	}

	// Token is expired or about to expire, refresh it
	newToken, err := tm.refreshAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	tm.token = newToken

	// Cache the new token
	if err := tm.cacheToken(); err != nil {
		// Log but don't fail - caching is optional
		fmt.Printf("Warning: failed to cache token: %v\n", err)
	}

	return tm.token, nil
}

// refreshAccessToken uses the refresh token to get a new access token
func (tm *TokenManager) refreshAccessToken(ctx context.Context) (*oauth2.Token, error) {
	// Ensure we have a refresh token
	if tm.refreshToken == "" && (tm.token == nil || tm.token.RefreshToken == "") {
		return nil, fmt.Errorf("no refresh token available")
	}

	// Use the refresh token from config or from existing token
	refreshToken := tm.refreshToken
	if refreshToken == "" && tm.token != nil {
		refreshToken = tm.token.RefreshToken
	}

	// Create a token with refresh token
	oldToken := &oauth2.Token{
		RefreshToken: refreshToken,
	}

	// Get token source and refresh
	tokenSource := tm.config.TokenSource(ctx, oldToken)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get new token: %w", err)
	}

	// Preserve refresh token if not returned
	if newToken.RefreshToken == "" {
		newToken.RefreshToken = refreshToken
	}

	return newToken, nil
}

// loadCachedToken loads token from cache file
func (tm *TokenManager) loadCachedToken() error {
	data, err := os.ReadFile(tm.cacheFile)
	if err != nil {
		return err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return err
	}

	// Ensure refresh token is set
	if token.RefreshToken == "" && tm.refreshToken != "" {
		token.RefreshToken = tm.refreshToken
	}

	tm.token = &token
	return nil
}

// cacheToken saves token to cache file
func (tm *TokenManager) cacheToken() error {
	// Create cache directory if it doesn't exist
	dir := filepath.Dir(tm.cacheFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.Marshal(tm.token)
	if err != nil {
		return err
	}

	return os.WriteFile(tm.cacheFile, data, 0600)
}

// ClearCache removes the cached token file
func (tm *TokenManager) ClearCache() error {
	return os.Remove(tm.cacheFile)
}

// TokenInfo returns information about the current token
func (tm *TokenManager) TokenInfo() (isValid bool, expiresIn time.Duration) {
	tm.tokenMu.RLock()
	defer tm.tokenMu.RUnlock()

	if tm.token == nil {
		return false, 0
	}

	isValid = tm.token.Valid()
	if isValid {
		expiresIn = time.Until(tm.token.Expiry)
	}

	return isValid, expiresIn
}
