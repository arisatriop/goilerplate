package utils

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateUUID_IsTimeOrderedV7(t *testing.T) {
	// Arrange / Act
	first := GenerateUUID()
	time.Sleep(2 * time.Millisecond)
	second := GenerateUUID()

	// Assert
	parsed, err := uuid.Parse(first)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), parsed.Version())
	assert.NotEqual(t, first, second)
	assert.Less(t, first, second, "later IDs sort after earlier ones")
}

// GenerateSecureToken is kept although nothing calls it yet: it is the crypto/rand primitive
// the password-reset flow needs, and it is the one function in this file that would be
// dangerous to get wrong. The generators that were deleted alongside it — GenerateRandomString
// and GenerateRandomNumberString — panicked on a rand failure under a comment reading "In real
// applications, handle the error appropriately".
func TestGenerateSecureToken_IsUnpredictableAndTheRightLength(t *testing.T) {
	// Act
	first, err := GenerateSecureToken(32)
	require.NoError(t, err)
	second, err := GenerateSecureToken(32)
	require.NoError(t, err)

	// Assert
	assert.NotEqual(t, first, second, "two tokens must never collide")

	decoded, err := base64.URLEncoding.DecodeString(first)
	require.NoError(t, err, "the token must be valid base64url, so it is safe in a URL")
	assert.Len(t, decoded, 32, "the caller asked for 32 bytes of entropy")
}

func TestGenerateSecureToken_ZeroLength(t *testing.T) {
	token, err := GenerateSecureToken(0)

	require.NoError(t, err)
	assert.Empty(t, token, "zero bytes of entropy is an empty token, not an error")
}
