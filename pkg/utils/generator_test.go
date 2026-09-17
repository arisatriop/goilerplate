package utils

import (
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
