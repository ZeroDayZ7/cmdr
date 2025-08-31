package crypto

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// EncryptFolderAES tworzy zip folderu i szyfruje go AES
func EncryptFolderAES(folderPath string, key []byte, outDir string) (string, string, error) {
	// Nazwa pliku wyjściowego .zip.enc
	folderName := filepath.Base(folderPath)
	encPath := filepath.Join(outDir, folderName+".zip.enc")
	keyPath := filepath.Join(outDir, "key.txt")

	// Tworzymy folder wyjściowy jeśli nie istnieje
	if err := os.MkdirAll(outDir, 0700); err != nil {
		return "", "", fmt.Errorf("failed to create output folder: %w", err)
	}

	// Zapisz klucz AES do key.txt
	if err := os.WriteFile(keyPath, []byte(fmt.Sprintf("%x", key)), 0600); err != nil {
		return "", "", fmt.Errorf("failed to write AES key: %w", err)
	}

	// Tworzymy zip w pamięci
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(folderPath, path)
		if err != nil {
			return err
		}

		fw, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		inFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer inFile.Close()

		_, err = io.Copy(fw, inFile)
		return err
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to create zip: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		return "", "", fmt.Errorf("failed to close zip: %w", err)
	}

	// Zapisz zip do pliku tymczasowego
	tmpZip := filepath.Join(outDir, folderName+".tmpzip")
	if err := os.WriteFile(tmpZip, buf.Bytes(), 0600); err != nil {
		return "", "", fmt.Errorf("failed to write temp zip: %w", err)
	}
	defer os.Remove(tmpZip)

	// Szyfrujemy zip
	if _, _, err := EncryptFileAES(tmpZip, key, outDir); err != nil {
		return "", "", fmt.Errorf("failed to encrypt zip: %w", err)
	}

	return encPath, keyPath, nil
}
