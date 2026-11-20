//go:build windows

package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"
)

func ApplyPlatformUpdate(currentPath, stagedPath string) error {
	if !writableDirectory(filepath.Dir(currentPath)) {
		return fmt.Errorf("current installation directory is not writable: %s", filepath.Dir(currentPath))
	}

	helper, err := os.Executable()
	if err != nil {
		return err
	}
	resultPath := stagedPath + ".result"
	process, err := os.StartProcess(helper, []string{helper, "update-helper", "--target", currentPath, "--staged", stagedPath, "--result", resultPath, "--pid", fmt.Sprintf("%d", os.Getpid())}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	if err != nil {
		return fmt.Errorf("start Windows update helper: %w", err)
	}
	_ = process.Release()

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if data, readErr := os.ReadFile(resultPath); readErr == nil {
			message := string(data)
			if message == "success" {
				return nil
			}
			return fmt.Errorf("Windows update helper failed: %s", message)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for Windows update helper")
}

func ApplyWindowsUpdate(target, staged string, pid int) error {
	if pid > 0 {
		process, err := os.FindProcess(pid)
		if err == nil {
			_, _ = process.Wait()
		}
	}

	backup := target + ".cmdr-backup"
	_ = os.Remove(backup)
	if err := os.Rename(target, backup); err != nil {
		return fmt.Errorf("backup existing executable: %w", err)
	}
	if err := replaceMoveFileEx(staged, target, true); err != nil {
		_ = os.Rename(backup, target)
		return fmt.Errorf("replace executable: %w", err)
	}
	info, err := os.Stat(target)
	if err != nil || info.IsDir() || info.Size() == 0 {
		_ = os.Remove(target)
		_ = os.Rename(backup, target)
		return fmt.Errorf("verify installed executable")
	}
	_ = os.Remove(backup)
	return nil
}

func ApplyWindowsUpdateHelper(target, staged, result string, pid int) error {
	if pid > 0 {
		if process, err := os.FindProcess(pid); err == nil {
			_, _ = process.Wait()
		}
	}
	err := ApplyWindowsUpdate(target, staged, pid)
	message := "success"
	if err != nil {
		message = err.Error()
	}
	_ = os.WriteFile(result, []byte(message), 0o600)
	return err
}

func replaceMoveFileEx(src, dst string, overwrite bool) error {
	srcPtr, err := windows.UTF16PtrFromString(src)
	if err != nil {
		return err
	}
	dstPtr, err := windows.UTF16PtrFromString(dst)
	if err != nil {
		return err
	}
	flags := uint32(windows.MOVEFILE_WRITE_THROUGH | windows.MOVEFILE_REPLACE_EXISTING)
	if !overwrite {
		flags = windows.MOVEFILE_WRITE_THROUGH
	}
	return windows.MoveFileEx(srcPtr, dstPtr, flags)
}

func writableDirectory(dir string) bool {
	probe, err := os.CreateTemp(dir, ".cmdr-write-test-*")
	if err != nil {
		return false
	}
	name := probe.Name()
	_ = probe.Close()
	_ = os.Remove(name)
	return true
}
