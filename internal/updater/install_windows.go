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
	resultPath := stagedPath + ".result"
	if err := ValidateUpdatePaths(currentPath, stagedPath, resultPath); err != nil {
		return err
	}
	if !writableDirectory(filepath.Dir(currentPath)) {
		return fmt.Errorf("current installation directory is not writable: %s", filepath.Dir(currentPath))
	}

	helper, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve cmdr executable path: %w", err)
	}
	helper, err = filepath.Abs(helper)
	if err != nil {
		return fmt.Errorf("resolve absolute cmdr executable path: %w", err)
	}
	if _, err := os.Stat(helper); err != nil {
		return fmt.Errorf("Windows update helper executable not found: %s: %w", helper, err)
	}

	process, err := os.StartProcess(helper, []string{helper, "update-helper",
		"--target", currentPath,
		"--staged", stagedPath,
		"--result", resultPath,
		"--pid", fmt.Sprintf("%d", os.Getpid())},
		&os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}, Dir: filepath.Dir(helper)})
	if err != nil {
		return fmt.Errorf("start Windows update helper %q: %w", helper, err)
	}
	_ = process.Release()
	return nil
}

func ApplyWindowsUpdate(target, staged string, pid int) error {
	if err := ValidateUpdatePaths(target, staged, target+".cmdr-update-result"); err != nil {
		return err
	}
	if pid <= 0 {
		return fmt.Errorf("invalid parent cmdr PID: %d", pid)
	}
	if err := waitForProcessExit(pid, 60*time.Second); err != nil {
		return fmt.Errorf("wait for cmdr process PID %d: %w", pid, err)
	}

	targetInfo, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("target executable is missing or inaccessible: %q: %w", target, err)
	}
	if targetInfo.IsDir() || targetInfo.Size() == 0 {
		return fmt.Errorf("invalid target executable: %q", target)
	}
	stagedInfo, err := os.Stat(staged)
	if err != nil {
		return fmt.Errorf("staged executable is missing or inaccessible: %q: %w", staged, err)
	}
	if stagedInfo.IsDir() || stagedInfo.Size() == 0 {
		return fmt.Errorf("invalid staged executable: %q", staged)
	}

	backup := target + ".cmdr-backup"
	_ = os.Remove(backup)
	if err := os.Rename(target, backup); err != nil {
		return fmt.Errorf("backup existing executable %q: %w", target, err)
	}

	if err := replaceMoveFileEx(staged, target, true); err != nil {
		restoreErr := os.Rename(backup, target)
		if restoreErr != nil {
			_ = os.Remove(staged)
			return fmt.Errorf("replace executable %q with %q failed: %v; rollback failed: %w", target, staged, err, restoreErr)
		}
		_ = os.Remove(staged)
		return fmt.Errorf("replace executable %q with %q failed: %w; original executable restored", target, staged, err)
	}

	info, err := os.Stat(target)
	if err != nil || info.IsDir() || info.Size() == 0 {
		_ = os.Remove(target)
		restoreErr := os.Rename(backup, target)
		if restoreErr != nil {
			return fmt.Errorf("verify installed executable %q failed; rollback failed: %w", target, restoreErr)
		}
		return fmt.Errorf("verify installed executable %q failed; original executable restored", target)
	}

	_ = os.Remove(backup)
	return nil
}

func ApplyWindowsUpdateHelper(target, staged, result string, pid int) error {
	if err := ValidateUpdatePaths(target, staged, result); err != nil {
		writeUpdateResult(result, err.Error())
		return err
	}

	err := ApplyWindowsUpdate(target, staged, pid)
	if err != nil {
		writeUpdateResult(result, err.Error())
		_ = os.Remove(staged)
		fmt.Fprintf(os.Stderr, "cmdr update failed: %v\n", err)
		return err
	}

	_ = os.Remove(staged)
	_ = os.Remove(result)
	fmt.Println("cmdr Windows update completed successfully.")
	return nil
}

func waitForProcessExit(pid int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
		if err != nil {
			// OpenProcess cannot find a handle after the parent process exits.
			return nil
		}
		status, waitErr := windows.WaitForSingleObject(handle, 250)
		_ = windows.CloseHandle(handle)
		if waitErr == nil && status == windows.WAIT_OBJECT_0 {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return fmt.Errorf("timed out after %s waiting for cmdr process PID %d", timeout, pid)
}

func writeUpdateResult(path, message string) {
	if path == "" {
		return
	}
	tmp := path + ".tmp"
	_ = os.Remove(tmp)
	if err := os.WriteFile(tmp, []byte(message), 0o600); err == nil {
		_ = os.Remove(path)
		_ = os.Rename(tmp, path)
	}
	_ = os.Remove(tmp)
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
