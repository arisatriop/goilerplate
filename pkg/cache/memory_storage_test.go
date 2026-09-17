package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_SetGetDelete(t *testing.T) {
	// Arrange
	storage := NewMemoryStorage()
	value := []byte("response")

	// Act
	require.NoError(t, storage.Set("key", value, time.Minute))
	value[0] = 'X' // callers may reuse their buffer

	// Assert
	got, err := storage.Get("key")
	require.NoError(t, err)
	assert.Equal(t, []byte("response"), got)

	got[0] = 'Y'
	again, _ := storage.Get("key")
	assert.Equal(t, []byte("response"), again, "returned slices are copies")

	require.NoError(t, storage.Delete("key"))
	missing, err := storage.Get("key")
	require.NoError(t, err)
	assert.Nil(t, missing)
}

func TestMemoryStorage_Expiry(t *testing.T) {
	// Arrange
	now := time.Now()
	storage := NewMemoryStorage()
	storage.now = func() time.Time { return now }
	require.NoError(t, storage.Set("short", []byte("a"), time.Second))
	require.NoError(t, storage.Set("forever", []byte("b"), 0))

	// Act
	now = now.Add(2 * time.Hour)

	// Assert
	short, _ := storage.Get("short")
	assert.Nil(t, short)
	forever, _ := storage.Get("forever")
	assert.Equal(t, []byte("b"), forever)
}

func TestMemoryStorage_SweepsExpiredEntriesOnWrite(t *testing.T) {
	now := time.Now()
	storage := NewMemoryStorage()
	storage.now = func() time.Time { return now }
	require.NoError(t, storage.Set("old", []byte("a"), time.Second))

	now = now.Add(2 * memorySweepInterval)
	require.NoError(t, storage.Set("new", []byte("b"), time.Minute))

	storage.mu.Lock()
	defer storage.mu.Unlock()
	assert.NotContains(t, storage.entries, "old")
	assert.Contains(t, storage.entries, "new")
}

func TestMemoryStorage_Reset(t *testing.T) {
	storage := NewMemoryStorage()
	require.NoError(t, storage.Set("a", []byte("1"), 0))

	require.NoError(t, storage.Reset())

	got, _ := storage.Get("a")
	assert.Nil(t, got)
	assert.NoError(t, storage.Close())
}
