package user_test

import (
	"strings"
	"testing"

	"goilerplate/internal/domain/user"
	"goilerplate/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_SetPassword_StoresAHashNotThePlaintext(t *testing.T) {
	// Arrange
	u := &user.User{Email: "someone@example.com"}

	// Act
	err := u.SetPassword("a-perfectly-fine-password")

	// Assert
	require.NoError(t, err)
	assert.NotEqual(t, "a-perfectly-fine-password", u.PasswordHash)
	assert.NoError(t, utils.CheckPassword("a-perfectly-fine-password", u.PasswordHash))
}

// The method used to be HashPassword(), which discarded bcrypt's error. A failure left
// PasswordHash as the empty string and the caller had no way to notice, so the account was
// persisted with a hash that had never been computed.
func TestUser_SetPassword_ReportsFailureAndLeavesTheFieldUntouched(t *testing.T) {
	// Arrange: bcrypt refuses anything past 72 bytes.
	u := &user.User{PasswordHash: "the-existing-hash"}

	// Act
	err := u.SetPassword(strings.Repeat("a", 73))

	// Assert
	require.Error(t, err)
	assert.Equal(t, "the-existing-hash", u.PasswordHash,
		"a failed hash must not blank the field")
}
