package user

import (
	"fmt"

	"goilerplate/pkg/utils"

	"github.com/google/uuid"
)

// User is the persisted account. PasswordHash only ever holds a bcrypt hash — the plaintext
// a caller starts from is passed to SetPassword and never assigned to the field directly.
type User struct {
	ID           uuid.UUID
	Name         string
	Phone        string
	Email        string
	Avatar       string
	IsActive     bool
	PasswordHash string
}

// SetPassword hashes plaintext and stores the result.
//
// It returns the error rather than swallowing it. The previous version discarded it, which left
// PasswordHash set to the empty string on a bcrypt failure and persisted an account nothing
// downstream could tell apart from a correctly hashed one.
func (u *User) SetPassword(plaintext string) error {
	hashed, err := utils.HashPassword(plaintext)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}
	u.PasswordHash = hashed
	return nil
}
