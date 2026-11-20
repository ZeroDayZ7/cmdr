//go:build !windows

package updater

import (
	"fmt"
	"os"
	"path/filepath"
)

func ApplyPlatformUpdate(currentPath, stagedPath string) error {
	dir := filepath.Dir(currentPath)
	probe, err := os.CreateTemp(dir, ".cmdr-write-test-*")
	if err != nil {
		return fmt.Errorf("current installation is not writable: %w", err)
	}
	probeName := probe.Name()
	_ = probe.Close()
	_ = os.Remove(probeName)

	backup := currentPath + ".cmdr-backup"
	_ = os.Remove(backup)
	if err := os.Rename(currentPath, backup); err != nil {
		return fmt.Errorf("backup current executable: %w", err)
	}
	if err := os.Chmod(stagedPath, 0o755); err != nil {
		_ = os.Rename(backup, currentPath)
		return fmt.Errorf("set executable permissions: %w", err)
	}
	if err := os.Rename(stagedPath, currentPath); err != nil {
		_ = os.Rename(backup, currentPath)
		return fmt.Errorf("replace executable: %w", err)
	}
	info, err := os.Stat(currentPath)
	if err != nil || info.IsDir() || info.Mode().Perm()&0o111 == 0 {
		_ = os.Remove(currentPath)
		_ = os.Rename(backup, currentPath)
		return fmt.Errorf("verify installed executable")
	}
	_ = os.Remove(backup)
	return nil
}
