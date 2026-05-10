package crypto

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func DecryptFolderAES(encZipPath string, key []byte, outDir string) error {

	tmpZip := encZipPath + ".tmpzip"
	if err := DecryptFileAES(encZipPath, key, tmpZip); err != nil {
		return fmt.Errorf("failed to decrypt zip: %w", err)
	}
	defer func() {
		_ = os.Remove(tmpZip)
	}()

	r, err := zip.OpenReader(tmpZip)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {

		err := func() error {
			fpath := filepath.Join(outDir, f.Name)

			if f.FileInfo().IsDir() {
				if err := os.MkdirAll(fpath, 0700); err != nil {
					return fmt.Errorf("failed to create directory: %w", err)
				}
				return nil
			}

			if err := os.MkdirAll(filepath.Dir(fpath), 0700); err != nil {
				return fmt.Errorf("failed to create parent directory: %w", err)
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return fmt.Errorf("failed to open output file: %w", err)
			}
			defer outFile.Close()

			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("failed to open zip file content: %w", err)
			}
			defer rc.Close()

			if _, err = io.Copy(outFile, rc); err != nil {
				return fmt.Errorf("failed to copy content to file: %w", err)
			}

			return nil
		}()

		if err != nil {
			return fmt.Errorf("error extracting %s: %w", f.Name, err)
		}
	}

	return nil
}
