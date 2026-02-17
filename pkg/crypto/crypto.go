package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"goilerplate/pkg/utils"
	"io"
	"net/http"
)

// EncryptString encrypts a plaintext string using AES-256-GCM
func EncryptString(plainText string, key string) (string, error) {
	// Derive a 32-byte key from the provided key using SHA-256
	hash := sha256.Sum256([]byte(key))
	keyBytes := hash[:]

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// DecryptString decrypts an encrypted string using AES-256-GCM
func DecryptString(encrypted string, key string) (string, error) {
	// Derive a 32-byte key from the provided key using SHA-256
	hash := sha256.Sum256([]byte(key))
	keyBytes := hash[:]

	ciphertext, err := base64.URLEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", utils.Error(http.StatusNotFound, "tenant not found", err)
	}

	return string(plaintext), nil
}
