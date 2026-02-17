package crypto

import (
	"testing"
)

func TestEncryptDecryptString(t *testing.T) {
	key := "my-secret-key-12345"
	plainText := "Hello, World!"

	// Test encryption
	encrypted, err := EncryptString(plainText, key)
	if err != nil {
		t.Fatalf("EncryptString failed: %v", err)
	}

	if encrypted == "" {
		t.Fatal("Encrypted string is empty")
	}

	if encrypted == plainText {
		t.Fatal("Encrypted string should not be the same as plaintext")
	}

	// Test decryption
	decrypted, err := DecryptString(encrypted, key)
	if err != nil {
		t.Fatalf("DecryptString failed: %v", err)
	}

	if decrypted != plainText {
		t.Fatalf("Decrypted string does not match original. Expected: %s, Got: %s", plainText, decrypted)
	}
}

func TestEncryptStringWithDifferentKeys(t *testing.T) {
	key1 := "key-one"
	key2 := "key-two"
	plainText := "test message"

	encrypted1, err := EncryptString(plainText, key1)
	if err != nil {
		t.Fatalf("EncryptString with key1 failed: %v", err)
	}

	encrypted2, err := EncryptString(plainText, key2)
	if err != nil {
		t.Fatalf("EncryptString with key2 failed: %v", err)
	}

	// Same plaintext with different keys should produce different ciphertext
	if encrypted1 == encrypted2 {
		t.Fatal("Same plaintext encrypted with different keys should produce different ciphertext")
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	correctKey := "correct-key"
	wrongKey := "wrong-key"
	plainText := "secret message"

	encrypted, err := EncryptString(plainText, correctKey)
	if err != nil {
		t.Fatalf("EncryptString failed: %v", err)
	}

	// Try to decrypt with wrong key
	_, err = DecryptString(encrypted, wrongKey)
	if err == nil {
		t.Fatal("DecryptString should fail with wrong key")
	}
}

func TestEncryptStringEmptyPlaintext(t *testing.T) {
	key := "test-key"
	plainText := ""

	encrypted, err := EncryptString(plainText, key)
	if err != nil {
		t.Fatalf("EncryptString with empty plaintext failed: %v", err)
	}

	decrypted, err := DecryptString(encrypted, key)
	if err != nil {
		t.Fatalf("DecryptString failed: %v", err)
	}

	if decrypted != plainText {
		t.Fatalf("Decrypted empty string does not match. Expected: '%s', Got: '%s'", plainText, decrypted)
	}
}

func TestEncryptStringUUID(t *testing.T) {
	key := "uuid-encryption-key"
	uuid := "550e8400-e29b-41d4-a716-446655440000"

	encrypted, err := EncryptString(uuid, key)
	if err != nil {
		t.Fatalf("EncryptString with UUID failed: %v", err)
	}

	decrypted, err := DecryptString(encrypted, key)
	if err != nil {
		t.Fatalf("DecryptString failed: %v", err)
	}

	if decrypted != uuid {
		t.Fatalf("Decrypted UUID does not match. Expected: %s, Got: %s", uuid, decrypted)
	}
}

func TestDecryptInvalidBase64(t *testing.T) {
	key := "test-key"
	invalidEncrypted := "not-valid-base64!@#$%"

	_, err := DecryptString(invalidEncrypted, key)
	if err == nil {
		t.Fatal("DecryptString should fail with invalid base64")
	}
}

func TestDecryptShortCiphertext(t *testing.T) {
	key := "test-key"
	shortEncrypted := "YWJj" // "abc" in base64, too short for valid ciphertext

	_, err := DecryptString(shortEncrypted, key)
	if err == nil {
		t.Fatal("DecryptString should fail with ciphertext too short")
	}
}

func TestEncryptStringConsistency(t *testing.T) {
	key := "consistency-key"
	plainText := "test data"

	// Encrypt the same plaintext twice
	encrypted1, err := EncryptString(plainText, key)
	if err != nil {
		t.Fatalf("First EncryptString failed: %v", err)
	}

	encrypted2, err := EncryptString(plainText, key)
	if err != nil {
		t.Fatalf("Second EncryptString failed: %v", err)
	}

	// Due to random nonce, encrypting the same plaintext twice should produce different ciphertext
	if encrypted1 == encrypted2 {
		t.Fatal("Encrypting same plaintext twice should produce different ciphertext (due to random nonce)")
	}

	// But both should decrypt to the same plaintext
	decrypted1, err := DecryptString(encrypted1, key)
	if err != nil {
		t.Fatalf("First DecryptString failed: %v", err)
	}

	decrypted2, err := DecryptString(encrypted2, key)
	if err != nil {
		t.Fatalf("Second DecryptString failed: %v", err)
	}

	if decrypted1 != plainText || decrypted2 != plainText {
		t.Fatal("Both decryptions should return the original plaintext")
	}
}
