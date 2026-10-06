// Package hash provides the one way tokens are hashed before they are stored or compared.
// Keeping it in one place means a change of algorithm cannot leave two call sites disagreeing.
package hash

import (
	"crypto/hmac"
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

// Keyed returns the hex-encoded HMAC-SHA256 of message under key. It is for low-entropy values
// such as a 6-digit code: an unkeyed digest of one is reversed by hashing all million
// candidates, while this cannot be reversed without the key.
func Keyed(key []byte, message string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(message)) // hash.Hash.Write never returns an error
	return hex.EncodeToString(mac.Sum(nil))
}

// KeyedEqual compares message against a stored Keyed hash in constant time.
func KeyedEqual(key []byte, message, storedHash string) bool {
	return subtle.ConstantTimeCompare([]byte(Keyed(key, message)), []byte(storedHash)) == 1
}
