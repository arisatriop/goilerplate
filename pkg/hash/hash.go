// Package hash provides the one way tokens are hashed before they are stored or compared.
// Keeping it in one place means a change of algorithm cannot leave two call sites disagreeing.
package hash

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

// Token returns the hex-encoded SHA-256 of a token. It is meant for high-entropy values
// (JWTs, random tokens) that need a lookup key or a stored fingerprint — never for
// passwords, which require a slow KDF such as bcrypt.
func Token(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// Equal compares a token against a stored hash in constant time.
func Equal(token, storedHash string) bool {
	return subtle.ConstantTimeCompare([]byte(Token(token)), []byte(storedHash)) == 1
}
