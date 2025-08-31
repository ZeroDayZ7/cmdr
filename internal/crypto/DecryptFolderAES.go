package crypto

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// DecryptFolderAES decrypts a .zip.enc file and extracts it to outDir
func DecryptFolderAES(encZipPath string, key []byte, outDir string) error {
	// Decrypt zip to temp file
	tmpZip := encZipPath + ".tmpzip"
	if err := DecryptFileAES(encZipPath, key, tmpZip); err != nil {
		return fmt.Errorf("failed to decrypt zip: %w", err)
	}
	defer os.Remove(tmpZip)

	// Open decrypted zip
	r, err := zip.OpenReader(tmpZip)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(outDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0700)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0700); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
