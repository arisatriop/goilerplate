package hash_test

import (
	"testing"

	"goilerplate/pkg/hash"

	"github.com/stretchr/testify/assert"
)

func TestKeyed(t *testing.T) {
	t.Parallel()
	key := []byte("a-server-side-secret-of-32-bytes!")

	stored := hash.Keyed(key, "token-id:123456")

	assert.Len(t, stored, 64, "hex-encoded SHA-256")
	assert.NotEqual(t, hash.Token("token-id:123456"), stored, "keyed differs from the plain digest")
	assert.NotEqual(t, stored, hash.Keyed([]byte("another-secret-of-thirty-two-byte"), "token-id:123456"),
		"without the key the hash cannot be reproduced")
	assert.True(t, hash.KeyedEqual(key, "token-id:123456", stored))
	assert.False(t, hash.KeyedEqual(key, "token-id:123457", stored))
	assert.False(t, hash.KeyedEqual(key, "token-id:123456", ""))
}
