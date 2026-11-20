//go:build windows

package updater

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows"
)

var openUpdateProcess = windows.OpenProcess
var waitForUpdateProcess = windows.WaitForSingleObject
var closeUpdateProcessHandle = windows.CloseHandle
var replaceUpdateExecutable = replaceMoveFileEx
var renameUpdateFile = os.Rename
var waitUpdateProcessExit = waitForProcessExit

func ApplyPlatformUpdate(currentPath, stagedPath string) error {
	resultPath := currentPath + ".cmdr-update.result"
	if err := ValidateUpdatePaths(currentPath, stagedPath, resultPath); err != nil {
		return err
	}
	if !writableDirectory(filepath.Dir(currentPath)) {
		return fmt.Errorf("current installation directory is not writable: %s", filepath.Dir(currentPath))
	}

	helper, err := os.Executable()
	// helper, err := filepath.Abs(helper)
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

	writeUpdateResult(resultPath, "pending")
	process, err := os.StartProcess(helper, []string{helper, "update-helper",
		"--target", currentPath,
		"--staged", stagedPath,
		"--result", resultPath,
		"--pid", fmt.Sprintf("%d", os.Getpid())},
		&os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}, Dir: filepath.Dir(helper)})
	if err != nil {
		message := fmt.Sprintf("error: start Windows update helper %q: %v", helper, err)
		writeUpdateResult(resultPath, message)
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

	if err := waitUpdateProcessExit(pid, 60*time.Second); err != nil {
		return fmt.Errorf("wait for cmdr process PID %d: %w", pid, err)
	}

	backup, err := createUniqueBackupPath(target)
	if err != nil {
		return err
	}
	if err := renameUpdateFile(target, backup); err != nil {
		return fmt.Errorf("backup existing executable %q: %w", target, err)
	}

	if err := replaceUpdateExecutable(staged, target, true); err != nil {
		restoreErr := renameUpdateFile(backup, target)
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
		restoreErr := renameUpdateFile(backup, target)
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
		writeUpdateResult(result, "error: "+err.Error())
		return err
	}

	err := ApplyWindowsUpdate(target, staged, pid)
	if err != nil {
		writeUpdateResult(result, "error: "+err.Error())
		_ = os.Remove(staged)
		fmt.Fprintf(os.Stderr, "cmdr update failed: %v\n", err)
		return err
	}

	_ = os.Remove(staged)
	writeUpdateResult(result, "success")
	fmt.Println("cmdr Windows update completed successfully.")
	return nil
}

func waitForProcessExit(pid int, timeout time.Duration) error {
	if pid <= 0 {
		return fmt.Errorf("invalid parent cmdr PID: %d", pid)
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		handle, err := openUpdateProcess(windows.SYNCHRONIZE, false, uint32(pid))
		if err != nil {
			if errors.Is(err, windows.ERROR_NOT_FOUND) {
				return nil
			}
			if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
				return fmt.Errorf("invalid parent cmdr PID %d", pid)
			}
			if errors.Is(err, windows.ERROR_INVALID_HANDLE) {
				return fmt.Errorf("invalid process handle while checking parent cmdr PID %d: %w", pid, err)
			}
			if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
				return fmt.Errorf("access denied while opening parent cmdr PID %d: %w", pid, err)
			}
			return fmt.Errorf("unable to open parent cmdr PID %d: %w", pid, err)
		}

		status, waitErr := waitForUpdateProcess(handle, 250)
		_ = closeUpdateProcessHandle(handle)
		if waitErr == nil && status == windows.WAIT_OBJECT_0 {
			return nil
		}
		if waitErr != nil && waitErr != windows.ERROR_TIMEOUT {
			return fmt.Errorf("wait for parent cmdr PID %d failed: %w", pid, waitErr)
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

func createUniqueBackupPath(target string) (string, error) {
	dir := filepath.Dir(target)
	base := filepath.Base(target)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	timestamp := time.Now().Format("20060102150405")
	backup := filepath.Join(dir, fmt.Sprintf("%s-%s%s", name, timestamp, ext))

	i := 1
	for {
		if _, err := os.Stat(backup); os.IsNotExist(err) {
			return backup, nil
		}
		backup = filepath.Join(dir, fmt.Sprintf("%s-%s-%d%s", name, timestamp, i, ext))
		i++
	}
}
