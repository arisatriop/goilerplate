package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost
	DefaultCost = 12
	// TokenLength is the length of random tokens
	TokenLength = 32
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword compares a password with its hash
func CheckPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}
	return nil
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateRefreshToken generates a secure refresh token
func GenerateRefreshToken() (string, error) {
	return GenerateSecureToken(TokenLength)
}

// GenerateVerificationToken generates a secure verification token
func GenerateVerificationToken() (string, error) {
	return GenerateSecureToken(TokenLength)
}

// HashToken creates a SHA256 hash of a token for secure storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// VerifyToken verifies a token against its hash using constant-time comparison
func VerifyToken(token, hash string) bool {
	tokenHash := HashToken(token)
	return subtle.ConstantTimeCompare([]byte(tokenHash), []byte(hash)) == 1
}
