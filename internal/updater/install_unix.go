//go:build !windows

package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var unixRename = os.Rename
var unixRemove = os.Remove
var unixChmod = os.Chmod

func ApplyPlatformUpdate(currentPath, stagedPath string) error {
	dir := filepath.Dir(currentPath)
	probe, err := os.CreateTemp(dir, ".cmdr-write-test-*")
	if err != nil {
		return fmt.Errorf("current installation is not writable: %w", err)
	}
	probeName := probe.Name()
	_ = probe.Close()
	_ = unixRemove(probeName)

	backup, err := createUniqueUnixBackupPath(currentPath)
	if err != nil {
		return err
	}
	if err := unixRename(currentPath, backup); err != nil {
		return fmt.Errorf("backup existing executable %q: %w", currentPath, err)
	}

	if err := unixChmod(stagedPath, 0o755); err != nil {
		return rollbackUnixUpdate(backup, currentPath,
			fmt.Errorf("set executable permissions on staged file %q: %w", stagedPath, err))
	}
	if err := unixRename(stagedPath, currentPath); err != nil {
		return rollbackUnixUpdate(backup, currentPath,
			fmt.Errorf("replace executable %q: %w", currentPath, err))
	}

	info, err := os.Stat(currentPath)
	if err != nil || info.IsDir() || info.Mode().Perm()&0o111 == 0 {
		verifyErr := fmt.Errorf("verify installed executable %q", currentPath)
		if err != nil {
			verifyErr = fmt.Errorf("verify installed executable %q: %w", currentPath, err)
		}
		return rollbackUnixUpdate(backup, currentPath, verifyErr)
	}

	if err := unixRemove(backup); err != nil {
		return fmt.Errorf("updated executable installed successfully, but could not remove backup %q: %w", backup, err)
	}
	return nil
}

func rollbackUnixUpdate(backup, currentPath string, primary error) error {
	if _, err := os.Lstat(currentPath); err == nil {
		if removeErr := unixRemove(currentPath); removeErr != nil {
			return fmt.Errorf("%v; rollback failed: could not remove invalid target %q: %w; backup preserved at %q", primary, currentPath, removeErr, backup)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("%v; rollback failed: could not inspect target %q: %w; backup preserved at %q", primary, currentPath, err, backup)
	}

	if err := unixRename(backup, currentPath); err != nil {
		return fmt.Errorf("%v; rollback failed: could not restore backup %q to %q: %w", primary, backup, currentPath, err)
	}
	return fmt.Errorf("%v; original executable restored", primary)
}

func createUniqueUnixBackupPath(target string) (string, error) {
	dir := filepath.Dir(target)
	base := filepath.Base(target)
	stamp := time.Now().Format("20060102150405.000000000")
	for i := 0; i < 10000; i++ {
		backup := filepath.Join(dir, fmt.Sprintf(".%s.cmdr-backup-%s", base, stamp))
		if i > 0 {
			backup = filepath.Join(dir, fmt.Sprintf(".%s.cmdr-backup-%s-%d", base, stamp, i))
		}
		if _, err := os.Lstat(backup); os.IsNotExist(err) {
			return backup, nil
		} else if err != nil {
			return "", fmt.Errorf("check backup path %q: %w", backup, err)
		}
	}
	return "", fmt.Errorf("could not allocate unique backup path for %q", target)
}
