package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func EncryptFileAES(inputPath string, key []byte, outDir string) (string, string, error) {
	if err := os.MkdirAll(outDir, 0700); err != nil {
		return "", "", fmt.Errorf("failed to create folder: %w", err)
	}

	fileName := filepath.Base(inputPath)
	encPath := filepath.Join(outDir, fileName+".enc")
	keyPath := filepath.Join(outDir, "key.txt")

	if err := os.WriteFile(keyPath, []byte(hex.EncodeToString(key)), 0600); err != nil {
		return "", "", fmt.Errorf("failed to write key: %w", err)
	}

	plainData, err := os.ReadFile(inputPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read input file: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", fmt.Errorf("failed to create AEAD: %w", err)
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aead.Seal(nonce, nonce, plainData, nil)

	if err := os.WriteFile(encPath, ciphertext, 0600); err != nil {
		return "", "", fmt.Errorf("failed to write encrypted file: %w", err)
	}

	return encPath, keyPath, nil
}
