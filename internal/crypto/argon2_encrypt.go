// internal/crypto/argon2_encrypt.go
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// EncryptFileArgon2 szyfruje plik za pomocą hasła (Argon2id + AES-GCM)
func EncryptFileArgon2(inputPath, password, outDir string) (string, error) {
	if err := os.MkdirAll(outDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create output folder: %w", err)
	}

	params := DefaultArgon2Params()
	key, salt, err := DeriveKeyArgon2id(password, params)
	if err != nil {
		return "", err
	}

	plainData, err := os.ReadFile(inputPath)
	if err != nil {
		return "", fmt.Errorf("failed to read input file: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Szyfrujemy dane (nonce na początku danych szyfrowanych)
	ciphertext := aead.Seal(nonce, nonce, plainData, nil)

	// Struktura pliku wyjściowego: [16 B salt][nonce + ciphertext]
	finalPayload := append(salt, ciphertext...)

	fileName := filepath.Base(inputPath)
	encPath := filepath.Join(outDir, fileName+".enc")

	if err := os.WriteFile(encPath, finalPayload, 0600); err != nil {
		return "", fmt.Errorf("failed to write encrypted file: %w", err)
	}

	return encPath, nil
}

// DecryptFileArgon2 odszyfrowuje plik za pomocą hasła (Argon2id + AES-GCM)
func DecryptFileArgon2(inputPath, password, outPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}

	params := DefaultArgon2Params()
	saltLen := int(params.SaltLen)

	if len(data) < saltLen {
		return fmt.Errorf("invalid encrypted file format: file too short")
	}

	salt := data[:saltLen]
	encryptedData := data[saltLen:]

	key := DeriveKeyWithSaltArgon2id(password, salt, params)

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aead.NonceSize()
	if len(encryptedData) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decryption failed (invalid password or corrupted file): %w", err)
	}

	if err := os.WriteFile(outPath, plaintext, 0600); err != nil {
		return fmt.Errorf("failed to write decrypted file: %w", err)
	}

	return nil
}
