package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const MaxArchiveBytes int64 = 128 << 20

type DownloadProgressFunc func(downloaded, total int64, elapsed time.Duration)

func DownloadToFile(url, destination string, maxBytes int64) error {
	return DownloadToFileWithProgress(url, destination, maxBytes, nil)
}

func DownloadToFileWithProgress(url, destination string, maxBytes int64, onProgress DownloadProgressFunc) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "cmdr-updater")
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}
	if resp.ContentLength >= 0 && resp.ContentLength > maxBytes {
		return fmt.Errorf("download exceeds size limit: %d bytes", resp.ContentLength)
	}

	out, err := os.CreateTemp(filepath.Dir(destination), ".download-*.tmp")
	if err != nil {
		return err
	}
	tmp := out.Name()
	cleanup := func() { _ = out.Close(); _ = os.Remove(tmp) }
	started := time.Now()
	limited := io.LimitReader(resp.Body, maxBytes+1)
	buffer := make([]byte, 32*1024)
	var written int64
	var lastProgress time.Time
	for {
		read, readErr := limited.Read(buffer)
		if read > 0 {
			writtenBytes, writeErr := out.Write(buffer[:read])
			written += int64(writtenBytes)
			if writeErr != nil {
				cleanup()
				return fmt.Errorf("write download: %w", writeErr)
			}
			now := time.Now()
			if onProgress != nil && (lastProgress.IsZero() || now.Sub(lastProgress) >= progressRefreshInterval) {
				lastProgress = now
				onProgress(written, resp.ContentLength, now.Sub(started))
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			cleanup()
			return fmt.Errorf("download interrupted: %w", readErr)
		}
	}
	if onProgress != nil {
		onProgress(written, resp.ContentLength, time.Since(started))
	}
	if written > maxBytes {
		cleanup()
		return fmt.Errorf("download exceeds size limit")
	}
	if err := out.Close(); err != nil {
		cleanup()
		return err
	}
	if err := os.Rename(tmp, destination); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func ExtractArchive(archivePath, destination, expectedName, platform string) (string, error) {
	if err := os.MkdirAll(destination, 0o700); err != nil {
		return "", err
	}
	var binaryName string
	if platform == "windows" {
		binaryName = "cmdr.exe"
	} else {
		binaryName = "cmdr"
	}

	if strings.HasSuffix(strings.ToLower(expectedName), ".zip") {
		return extractZip(archivePath, destination, binaryName)
	}
	if strings.HasSuffix(strings.ToLower(expectedName), ".tar.gz") {
		return extractTarGz(archivePath, destination, binaryName)
	}
	return "", fmt.Errorf("unsupported archive artifact: %s", expectedName)
}

func extractZip(archivePath, destination, binaryName string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("open ZIP archive: %w", err)
	}
	defer reader.Close()
	var found string
	for _, file := range reader.File {
		if file.FileInfo().Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("symlink in archive is not allowed: %s", file.Name)
		}
		name, err := safeArchivePath(destination, file.Name)
		if err != nil {
			return "", err
		}
		base := filepath.Base(file.Name)
		if base == binaryName {
			if found != "" {
				return "", fmt.Errorf("archive contains duplicate %s", binaryName)
			}
			r, err := file.Open()
			if err != nil {
				return "", err
			}
			found, err = copyArchiveFile(r, name, 0o755)
			r.Close()
			if err != nil {
				return "", err
			}
			continue
		}
		if base != "" && (base == "README.md" || base == "LICENSE") {
			continue
		}
		if base == "" {
			continue
		}
		return "", fmt.Errorf("unexpected file in archive: %s", file.Name)
	}
	if found == "" {
		return "", fmt.Errorf("archive does not contain %s", binaryName)
	}
	return found, nil
}

func extractTarGz(archivePath, destination, binaryName string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("open tar.gz archive: %w", err)
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	var found string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("read tar archive: %w", err)
		}
		if header.Typeflag == tar.TypeSymlink || header.Typeflag == tar.TypeLink {
			return "", fmt.Errorf("links in archive are not allowed: %s", header.Name)
		}
		name, err := safeArchivePath(destination, header.Name)
		if err != nil {
			return "", err
		}
		base := filepath.Base(header.Name)
		if base == binaryName {
			if found != "" {
				return "", fmt.Errorf("archive contains duplicate %s", binaryName)
			}
			found, err = copyArchiveFile(tarReader, name, 0o755)
			if err != nil {
				return "", err
			}
			continue
		}
		if base == "README.md" || base == "LICENSE" {
			continue
		}
		if header.Typeflag == tar.TypeDir {
			continue
		}
		return "", fmt.Errorf("unexpected file in archive: %s", header.Name)
	}
	if found == "" {
		return "", fmt.Errorf("archive does not contain %s", binaryName)
	}
	return found, nil
}

func safeArchivePath(destination, name string) (string, error) {
	clean := filepath.Clean(name)
	slashName := strings.ReplaceAll(name, "\\", "/")
	if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(slashName, "/") || strings.Contains(slashName, ":") || strings.Contains(slashName, "/../") || strings.HasPrefix(slashName, "../") || slashName == ".." {
		return "", fmt.Errorf("unsafe archive path: %s", name)
	}
	full := filepath.Join(destination, clean)
	rel, err := filepath.Rel(destination, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe archive path: %s", name)
	}
	return full, nil
}

func copyArchiveFile(src io.Reader, destination string, mode os.FileMode) (string, error) {
	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return "", err
	}
	written, err := io.Copy(out, io.LimitReader(src, MaxArchiveBytes+1))
	closeErr := out.Close()
	if err != nil {
		_ = os.Remove(destination)
		return "", err
	}
	if closeErr != nil {
		return "", closeErr
	}
	if written > MaxArchiveBytes {
		_ = os.Remove(destination)
		return "", fmt.Errorf("binary in archive exceeds size limit")
	}
	return destination, nil
}

func CreateUpdateTempDir() (string, error) {
	return os.MkdirTemp("", "cmdr-update-")
}

func CleanUpdateTempDir(path string) {
	_ = os.RemoveAll(path)
}
