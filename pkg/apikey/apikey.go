// Package apikey verifies partner API keys.
//
// A key is a bearer secret: whoever holds it is the partner. That makes two things matter more
// than they would for a user password. The comparison must not leak how much of a guess was
// right, and the key itself must not survive the comparison — not in a context value, not in a
// log line, not as a cache key.
package apikey

import (
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"goilerplate/pkg/hash"
)

// HashPrefix marks a configured value that is already a SHA-256 digest rather than the key
// itself. Deployments that would rather not keep plaintext keys in a config file store
// "sha256:<64 hex chars>" instead; the server never needs the original.
const HashPrefix = "sha256:"

// digestLength is the hex length of a SHA-256 digest.
const digestLength = 64

// entry is one configured partner. Only the digest is kept: once the registry is built there is
// no copy of the plaintext key anywhere in the process.
type entry struct {
	name   string
	digest string
}

// Registry verifies presented keys against the configured partners.
//
// Entries are a slice rather than a map because the lookup deliberately visits every one of
// them. A map lookup would return as soon as it found a match, and how long that took would
// depend on which partner presented the key.
type Registry struct {
	entries []entry
}

// NewRegistry builds a registry from the configured name → key (or name → "sha256:...") pairs.
//
// A malformed digest is kept as-is rather than rejected here: it simply never matches, so the
// runtime behaviour is fail-closed. Config validation reports it at startup, which is where a
// typo should surface — not as a partner that silently stops being able to authenticate.
func NewRegistry(keys map[string]string) *Registry {
	entries := make([]entry, 0, len(keys))
	for name, value := range keys {
		entries = append(entries, entry{name: name, digest: Digest(value)})
	}

	return &Registry{entries: entries}
}

// Digest returns the stored form of a configured value: the value itself when it is already a
// "sha256:" digest, otherwise the digest of the plaintext key.
func Digest(configured string) string {
	if rest, found := strings.CutPrefix(configured, HashPrefix); found {
		return strings.ToLower(strings.TrimSpace(rest))
	}
	return hash.Token(configured)
}

// Lookup returns the partner that owns the presented key.
//
// The key is hashed first and the comparison is over two digests, which are always the same
// length. Comparing the raw strings would return early on a length mismatch and so reveal how
// long the real key is. Every entry is visited even after a match, so the time taken does not
// depend on which partner matched, or on whether any did.
func (r *Registry) Lookup(presented string) (string, bool) {
	if presented == "" {
		return "", false
	}

	digest := hash.Token(presented)

	name := ""
	matched := 0
	for _, e := range r.entries {
		if subtle.ConstantTimeCompare([]byte(digest), []byte(e.digest)) == 1 {
			name = e.name
			matched = 1
		}
	}

	if matched == 0 {
		return "", false
	}
	return name, true
}

// Len reports how many partners are configured.
func (r *Registry) Len() int { return len(r.entries) }

// ValidateConfigured reports why a configured value cannot be used. It exists so a typo in a
// digest is reported at startup rather than becoming a partner who can never authenticate.
func ValidateConfigured(value string) error {
	rest, found := strings.CutPrefix(value, HashPrefix)
	if !found {
		return nil
	}

	rest = strings.TrimSpace(rest)
	if len(rest) != digestLength {
		return fmt.Errorf("must be %s followed by %d hex characters, got %d", HashPrefix, digestLength, len(rest))
	}
	if _, err := hex.DecodeString(rest); err != nil {
		return fmt.Errorf("must be %s followed by hex characters", HashPrefix)
	}

	return nil
}

// IsHashed reports whether a configured value is a digest rather than a plaintext key.
func IsHashed(value string) bool {
	return strings.HasPrefix(value, HashPrefix)
}
