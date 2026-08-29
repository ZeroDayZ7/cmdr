// internal/crypto/random.go
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// GenerateRandomAESKey generates a random AES key of given length in bytes
func GenerateRandomAESKey(length int) ([]byte, error) {
	return GenerateRandomBytes(length)
}

// GenerateRandomBytes generates a random byte slice of specified length
func GenerateRandomBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, fmt.Errorf("length must be positive")
	}
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return bytes, nil
}

// GenerateRandomString generates a random alphanumeric string
func GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random string: %w", err)
		}
		result[i] = letters[num.Int64()]
	}
	return string(result), nil
}

// GenerateRandomPassword generates a secure password with letters, numbers, and symbols
func GenerateRandomPassword(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:,.<>?"
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random password: %w", err)
		}
		result[i] = chars[num.Int64()]
	}
	return string(result), nil
}

// GenerateRandomNumber generates a random number with up to maxDigits digits
func GenerateRandomNumber(maxDigits int) (int64, error) {
	if maxDigits <= 0 || maxDigits > 18 {
		return 0, fmt.Errorf("maxDigits must be between 1 and 18")
	}
	limit := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(maxDigits)), nil)
	num, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return 0, fmt.Errorf("failed to generate random number: %w", err)
	}
	return num.Int64(), nil
}

// GenerateRandomHex generates a random hex-encoded string of specified byte length
func GenerateRandomHex(length int) (string, error) {
	bytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateRandomHash generates a SHA-256 hash of random bytes
func GenerateRandomHash(length int) (string, error) {
	bytes, err := GenerateRandomBytes(length)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:]), nil
}
