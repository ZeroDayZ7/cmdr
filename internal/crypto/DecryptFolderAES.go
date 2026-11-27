package crypto

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func DecryptFolderAES(encZipPath string, key []byte, outDir string) (retErr error) {

	tmpZip := encZipPath + ".tmpzip"
	if err := DecryptFileAES(encZipPath, key, tmpZip); err != nil {
		return fmt.Errorf("failed to decrypt zip: %w", err)
	}
	defer func() {
		if err := os.Remove(tmpZip); err != nil && !os.IsNotExist(err) && retErr == nil {
			retErr = fmt.Errorf("remove temporary zip: %w", err)
		}
	}()

	r, err := zip.OpenReader(tmpZip)
	if err != nil {
		return fmt.Errorf("failed to open zip: %w", err)
	}
	defer func() {
		if err := r.Close(); err != nil && retErr == nil {
			retErr = fmt.Errorf("close zip reader: %w", err)
		}
	}()

	for _, f := range r.File {

		err := func() (retErr error) {
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
			defer func() {
				if err := outFile.Close(); err != nil && retErr == nil {
					retErr = fmt.Errorf("close extracted file: %w", err)
				}
			}()

			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("failed to open zip file content: %w", err)
			}
			defer func() {
				if err := rc.Close(); err != nil && retErr == nil {
					retErr = fmt.Errorf("close zip file content: %w", err)
				}
			}()

			if _, err = io.Copy(outFile, rc); err != nil {
				return fmt.Errorf("failed to copy content: %w", err)
			}
			return nil
		}()

		if err != nil {
			return fmt.Errorf("error extracting %s: %w", f.Name, err)
		}
	}

	return nil
}
