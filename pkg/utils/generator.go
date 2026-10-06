package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// DefaultCost is the default bcrypt cost.
const DefaultCost = 12

// passwordCost is the cost HashPassword uses. It is DefaultCost everywhere but in tests, which
// lower it through SetPasswordCostForTests.
var passwordCost atomic.Int64

func init() { passwordCost.Store(DefaultCost) }

// SetPasswordCostForTests sets the bcrypt cost HashPassword uses, and that SimulatePasswordCheck
// matches. It exists for TestMain: at DefaultCost one hash takes about two seconds under -race,
// which made the auth suites take minutes. Never call it from application code — the cost is
// what makes a leaked hash slow to crack.
func SetPasswordCostForTests(cost int) {
	passwordCost.Store(int64(cost))
}

// GenerateUUID generates a time-ordered UUIDv7 string, used for primary keys so new rows
// stay close together in B-tree indexes.
func GenerateUUID() string {
	return uuid.Must(uuid.NewV7()).String()
}

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), int(passwordCost.Load()))
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

// SimulatePasswordCheck spends the time a real password check would, against a hash that
// nothing matches. A login path with no password to verify — an unknown email, a locked account
// — calls it so that it takes as long as one that has, and response time cannot tell an
// attacker which addresses are registered.
//
// The hash is made with the same cost as HashPassword, at first use. It used to be a constant
// at cost 10 while passwords were hashed at 12, so an unknown email answered about four times
// faster than a wrong password: the timing difference this function exists to remove.
func SimulatePasswordCheck(password string) {
	_ = bcrypt.CompareHashAndPassword(dummyHash(), []byte(password))
}

var dummy struct {
	sync.Mutex
	cost int
	hash []byte
}

// dummyHash returns a bcrypt hash of a random value at the current password cost, regenerating
// it if the cost has changed since.
func dummyHash() []byte {
	cost := int(passwordCost.Load())

	dummy.Lock()
	defer dummy.Unlock()

	if dummy.hash == nil || dummy.cost != cost {
		secret := make([]byte, 32)
		_, _ = rand.Read(secret) // crypto/rand.Read never returns an error
		// GenerateFromPassword fails only for a cost out of range or input over 72 bytes,
		// neither of which this can produce.
		hash, _ := bcrypt.GenerateFromPassword(secret, cost)
		dummy.cost, dummy.hash = cost, hash
	}
	return dummy.hash
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
