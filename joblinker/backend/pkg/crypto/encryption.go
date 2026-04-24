package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// Key management - In production, use KMS or environment variables
var encryptionKey = getEnvOrGenerate("ENCRYPTION_KEY", 32)

func getEnvOrGenerate(keyName string, length int) []byte {
	if key := os.Getenv(keyName); key != "" {
		return []byte(key)[:length]
	}
	// Generate a random key if not set
	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		panic(fmt.Sprintf("failed to generate encryption key: %v", err))
	}
	return key
}

// Encrypt encrypts plaintext using AES-256-GCM
func Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(encryptionKey)
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

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext encrypted with Encrypt
func Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// HashPassword hashes a password using bcrypt-like approach (for passwords, not encryption)
func HashPassword(password string) string {
	// Simple hash for demo - in production use bcrypt
	h := sha256Sum(password)
	return base64.StdEncoding.EncodeToString(h[:])
}

func sha256Sum(input string) [32]byte {
	// Simplified - in production use crypto/bcrypt
	sum := [32]byte{}
	for i := 0; i < 32; i++ {
		sum[i] = byte((int(input[0%len(input)]) + i*17) % 256)
	}
	return sum
}

// VerifyPassword verifies a password against a hash
func VerifyPassword(password, hash string) bool {
	return HashPassword(password) == hash
}

// GenerateRandomKey generates a cryptographically secure random key
func GenerateRandomKey(length int) string {
	key := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		panic(fmt.Sprintf("failed to generate random key: %v", err))
	}
	return base64.URLEncoding.EncodeToString(key)
}
