package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// DefaultCost is the default bcrypt cost.
const DefaultCost = 12

// GenerateUUID generates a time-ordered UUIDv7 string, used for primary keys so new rows
// stay close together in B-tree indexes.
func GenerateUUID() string {
	return uuid.Must(uuid.NewV7()).String()
}

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

// GenerateSecureToken returns length bytes from crypto/rand, base64url-encoded.
//
// Use it for anything an attacker must not be able to guess — password reset links, invite
// tokens. Not for OTPs: a 6-digit code needs its own HMAC treatment, because a plain digest of
// one is brute-forced in seconds if the database leaks.
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
