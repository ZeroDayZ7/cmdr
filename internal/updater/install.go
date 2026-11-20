package updater

import (
	"fmt"
	"os"
	"path/filepath"
)

func ExecutablePath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err == nil {
		path = resolved
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return "", err
	}
	if info.IsDir() || !info.Mode().IsRegular() {
		return "", fmt.Errorf("executable is not a regular file: %s", absolute)
	}
	return filepath.Clean(absolute), nil
}

func PrepareStagedBinary(newBinary, targetDir string) (string, error) {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", err
	}
	out, err := os.CreateTemp(targetDir, ".cmdr-update-*.tmp")
	if err != nil {
		return "", err
	}
	staged := out.Name()
	out.Close()

	input, err := os.Open(newBinary)
	if err != nil {
		_ = os.Remove(staged)
		return "", err
	}
	defer input.Close()
	output, err := os.OpenFile(staged, os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		_ = os.Remove(staged)
		return "", err
	}
	if _, err := output.ReadFrom(input); err != nil {
		output.Close()
		_ = os.Remove(staged)
		return "", err
	}
	if err := output.Close(); err != nil {
		_ = os.Remove(staged)
		return "", err
	}
	return staged, nil
}
