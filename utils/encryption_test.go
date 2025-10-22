package utils

import (
	"os"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	testCases := []struct {
		name      string
		plaintext string
	}{
		{"Empty string", ""},
		{"Simple string", "hello world"},
		{"Environment variable", "DATABASE_URL=postgres://user:pass@localhost/db"},
		{"Complex value", "SECRET_KEY=this-is-a-very-long-secret-key-with-special-chars-123!@#$%^&*()"},
		{"Multi-line", "KEY1=value1\nKEY2=value2\nKEY3=value3"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test encryption
			encrypted, err := EncryptValue(tc.plaintext)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Empty string should return empty
			if tc.plaintext == "" && encrypted != "" {
				t.Errorf("Expected empty string for empty input, got %s", encrypted)
			}

			// Non-empty string should be encrypted (different from original)
			if tc.plaintext != "" && encrypted == tc.plaintext {
				t.Errorf("Encrypted value should be different from plaintext")
			}

			// Test decryption
			decrypted, err := DecryptValue(encrypted)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			// Decrypted should match original
			if decrypted != tc.plaintext {
				t.Errorf("Decrypted value does not match original. Expected: %s, Got: %s", tc.plaintext, decrypted)
			}
		})
	}
}

func TestEncryptionWithCustomKey(t *testing.T) {
	// Set a custom key
	os.Setenv("ORBIT_ENCRYPTION_KEY", "test-custom-key-for-testing")
	defer os.Unsetenv("ORBIT_ENCRYPTION_KEY")

	plaintext := "test-secret-value"

	encrypted, err := EncryptValue(plaintext)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decrypted, err := DecryptValue(encrypted)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Expected %s, got %s", plaintext, decrypted)
	}
}

func TestDecryptInvalidData(t *testing.T) {
	invalidCases := []string{
		"invalid-base64-data",
		"dGVzdA==", // Valid base64 but too short for cipher
		"",         // Empty string should return empty, not error
	}

	for _, invalid := range invalidCases {
		_, err := DecryptValue(invalid)
		if invalid == "" {
			// Empty string should not error
			if err != nil {
				t.Errorf("Empty string should not produce error, got: %v", err)
			}
		} else {
			// Invalid data should error
			if err == nil {
				t.Errorf("Expected error for invalid data: %s", invalid)
			}
		}
	}
}

func TestEncryptionDeterminism(t *testing.T) {
	plaintext := "test-value-for-determinism"

	// Encrypt the same value multiple times
	encrypted1, err1 := EncryptValue(plaintext)
	encrypted2, err2 := EncryptValue(plaintext)

	if err1 != nil || err2 != nil {
		t.Fatalf("Encryption failed: %v, %v", err1, err2)
	}

	// Due to random nonce, encrypted values should be different
	if encrypted1 == encrypted2 {
		t.Errorf("Multiple encryptions of the same value should produce different ciphertext due to random nonce")
	}

	// But both should decrypt to the same value
	decrypted1, err1 := DecryptValue(encrypted1)
	decrypted2, err2 := DecryptValue(encrypted2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Decryption failed: %v, %v", err1, err2)
	}

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Errorf("Both decrypted values should match original: %s, %s", decrypted1, decrypted2)
	}
}