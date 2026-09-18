package apikey

import (
	"strings"
	"testing"

	"goilerplate/pkg/hash"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const partnerKey = "partner-key-3f9a2c7e1b4d8a6f0c5e"

func TestRegistry_LookupResolvesThePartnerName(t *testing.T) {
	registry := NewRegistry(map[string]string{
		"acme":  partnerKey,
		"other": "another-key-8c1d4b7a2e6f9038aa51",
	})

	name, ok := registry.Lookup(partnerKey)

	require.True(t, ok)
	assert.Equal(t, "acme", name)
}

func TestRegistry_LookupRejectsUnknownAndEmptyKeys(t *testing.T) {
	registry := NewRegistry(map[string]string{"acme": partnerKey})

	for _, presented := range []string{
		"",
		"wrong-key",
		partnerKey + "x",               // a correct key with something appended
		partnerKey[:len(partnerKey)-1], // a correct prefix, one character short
		strings.ToUpper(partnerKey),
	} {
		name, ok := registry.Lookup(presented)
		assert.False(t, ok, "must not accept %q", presented)
		assert.Empty(t, name)
	}
}

// A deployment that would rather not keep plaintext keys in its config stores the digest. The
// server never needs the original, so both forms have to authenticate the same partner.
func TestRegistry_AcceptsPreHashedConfiguration(t *testing.T) {
	registry := NewRegistry(map[string]string{
		"acme": HashPrefix + hash.Token(partnerKey),
	})

	name, ok := registry.Lookup(partnerKey)

	require.True(t, ok)
	assert.Equal(t, "acme", name)
}

func TestRegistry_PreHashedConfigurationIsCaseInsensitive(t *testing.T) {
	registry := NewRegistry(map[string]string{
		"acme": HashPrefix + strings.ToUpper(hash.Token(partnerKey)),
	})

	_, ok := registry.Lookup(partnerKey)

	assert.True(t, ok, "a digest pasted in upper case is the same digest")
}

// A digest that cannot be right must reject every key rather than accept a convenient one.
// Startup validation is what tells the operator about it; the runtime only has to stay closed.
func TestRegistry_MalformedDigestMatchesNothing(t *testing.T) {
	registry := NewRegistry(map[string]string{"acme": HashPrefix + "not-a-digest"})

	for _, presented := range []string{partnerKey, "not-a-digest", HashPrefix + "not-a-digest", ""} {
		_, ok := registry.Lookup(presented)
		assert.False(t, ok, "must not accept %q against a malformed digest", presented)
	}
}

// With no partners configured nothing can authenticate — in particular not an empty header,
// which would otherwise be the one key that matches "no keys".
func TestRegistry_EmptyRegistryAcceptsNothing(t *testing.T) {
	registry := NewRegistry(nil)

	assert.Zero(t, registry.Len())

	_, ok := registry.Lookup("")
	assert.False(t, ok)

	_, ok = registry.Lookup(partnerKey)
	assert.False(t, ok)
}

// The registry must not keep the plaintext key alive: config is read once, and anything held
// after that is a credential sitting in memory for the life of the process for no reason.
func TestNewRegistry_KeepsOnlyDigests(t *testing.T) {
	registry := NewRegistry(map[string]string{"acme": partnerKey})

	require.Len(t, registry.entries, 1)
	assert.Equal(t, hash.Token(partnerKey), registry.entries[0].digest)
	assert.NotContains(t, registry.entries[0].digest, partnerKey)
}

func TestValidateConfigured(t *testing.T) {
	valid := hash.Token(partnerKey)

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"plaintext key is not checked here", partnerKey, false},
		{"a well formed digest", HashPrefix + valid, false},
		{"digest with surrounding space", HashPrefix + " " + valid + " ", false},
		{"digest too short", HashPrefix + valid[:63], true},
		{"digest too long", HashPrefix + valid + "a", true},
		{"digest with non-hex characters", HashPrefix + strings.Repeat("z", 64), true},
		{"prefix with nothing after it", HashPrefix, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateConfigured(tc.value)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestIsHashed(t *testing.T) {
	assert.True(t, IsHashed(HashPrefix+hash.Token(partnerKey)))
	assert.False(t, IsHashed(partnerKey))
	assert.False(t, IsHashed(""))
}
