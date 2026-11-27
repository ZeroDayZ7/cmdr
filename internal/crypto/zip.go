package crypto

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

// ZipFolder zips all files from folderPath into zipPath
func ZipFolder(folderPath, zipPath string) (retErr error) {
	outFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer func() {
		if err := outFile.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	zipWriter := zip.NewWriter(outFile)
	defer func() {
		if err := zipWriter.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	walkErr := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
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
		defer func() { _ = inFile.Close() }()

		_, err = io.Copy(fw, inFile)
		return err
	})
	return walkErr
}
