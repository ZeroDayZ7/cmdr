//go:build windows

package updater

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

func TestWriteUpdateResultAtomic(t *testing.T) {
	dir := t.TempDir()
	for _, status := range []string{"pending", "success", "error: replace failed"} {
		result := filepath.Join(dir, "update.result")
		if err := writeUpdateResult(result, status); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(result)
		if err != nil || string(data) != status {
			t.Fatalf("status = %q, %v", data, err)
		}
		if _, err := os.Stat(result + ".tmp"); !os.IsNotExist(err) {
			t.Fatal("temporary result file remains")
		}
	}
}

func TestWriteUpdateResultLegacyBody(t *testing.T) {
	dir := t.TempDir()
	result := filepath.Join(dir, "update.result")
	if err := writeUpdateResult(result, "replace failed"); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(result)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "replace failed" {
		t.Fatalf("result = %q", string(data))
	}
	if _, err := os.Stat(result + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("temporary result file remains")
	}
}

func TestApplyWindowsUpdateInvalidPID(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "cmdr.exe")
	staged := filepath.Join(dir, "cmdr.tmp")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staged, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := ApplyWindowsUpdate(target, staged, 0); err == nil {
		t.Fatal("expected invalid PID error")
	}
}

func TestApplyWindowsUpdateMissingStaged(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "cmdr.exe")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := ApplyWindowsUpdate(target, filepath.Join(dir, "missing.tmp"), os.Getpid()); err == nil {
		t.Fatal("expected missing staged file error")
	}
}

func TestCreateUniqueBackupPathPreservesExistingBackup(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "cmdr.exe")
	existing := filepath.Join(dir, "cmdr-20261120120000.exe")
	if err := os.WriteFile(existing, []byte("existing backup"), 0o755); err != nil {
		t.Fatal(err)
	}
	backup, err := createUniqueBackupPath(target)
	if err != nil {
		t.Fatal(err)
	}
	if backup == existing {
		t.Fatal("backup path collides with existing backup")
	}
	data, err := os.ReadFile(existing)
	if err != nil || string(data) != "existing backup" {
		t.Fatal("existing backup was modified")
	}
}

func TestApplyWindowsUpdateReplaceFailureRollsBack(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "cmdr.exe")
	staged := filepath.Join(dir, "staged.tmp")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staged, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}

	oldReplace := replaceUpdateExecutable
	replaceUpdateExecutable = func(_, _ string, _ bool) error { return os.ErrPermission }
	defer func() { replaceUpdateExecutable = oldReplace }()

	err := ApplyWindowsUpdate(target, staged, os.Getpid())
	if err == nil {
		t.Fatal("expected replacement failure")
	}
	data, readErr := os.ReadFile(target)
	if readErr != nil || string(data) != "old" {
		t.Fatal("target was not restored after failed replacement")
	}
}

func TestWaitForProcessExitInvalidPID(t *testing.T) {
	if err := waitForProcessExit(0, time.Second); err == nil {
		t.Fatal("expected invalid PID error")
	}
}

func TestApplyWindowsUpdateProcessWaitFailureDoesNotReplace(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "cmdr.exe")
	staged := filepath.Join(dir, "staged.tmp")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staged, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}

	oldWait := waitUpdateProcessExit
	waitUpdateProcessExit = func(int, time.Duration) error { return os.ErrPermission }
	defer func() { waitUpdateProcessExit = oldWait }()

	err := ApplyWindowsUpdate(target, staged, os.Getpid())
	if err == nil {
		t.Fatal("expected replacement failure")
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "old" {
		t.Fatal("target changed after process wait failure")
	}
}

func TestApplyWindowsUpdateFailedRollbackKeepsBackup(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "cmdr.exe")
	staged := filepath.Join(dir, "staged.tmp")
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staged, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}

	oldReplace := replaceUpdateExecutable
	oldRename := renameUpdateFile
	replaceUpdateExecutable = func(_, _ string, _ bool) error { return os.ErrPermission }
	restoreCalls := 0
	renameUpdateFile = func(from, to string) error {
		if restoreCalls > 0 {
			return os.ErrPermission
		}
		if filepath.Base(to) == "cmdr.exe" && filepath.Base(from) != "cmdr.exe" {
			restoreCalls++
			return os.ErrPermission
		}
		return oldRename(from, to)
	}
	oldWait := waitUpdateProcessExit
	waitUpdateProcessExit = func(int, time.Duration) error { return nil }
	defer func() {
		replaceUpdateExecutable = oldReplace
		renameUpdateFile = oldRename
		waitUpdateProcessExit = oldWait
	}()

	err := ApplyWindowsUpdate(target, staged, os.Getpid())
	if err == nil || !strings.Contains(err.Error(), "rollback failed") {
		t.Fatalf("expected rollback failure, got %v", err)
	}
}

func TestWaitForProcessExitParentExitedBeforeOpen(t *testing.T) {
	oldOpen := openUpdateProcess
	openUpdateProcess = func(uint32, bool, uint32) (windows.Handle, error) {
		return windows.InvalidHandle, windows.ERROR_INVALID_PARAMETER
	}
	defer func() { openUpdateProcess = oldOpen }()

	if err := waitForProcessExit(1234, time.Second); err != nil {
		t.Fatalf("expected exited parent to allow replacement, got %v", err)
	}
}

func TestWaitForProcessExitAccessDenied(t *testing.T) {
	oldOpen := openUpdateProcess
	openUpdateProcess = func(uint32, bool, uint32) (windows.Handle, error) {
		return windows.InvalidHandle, windows.ERROR_ACCESS_DENIED
	}
	defer func() { openUpdateProcess = oldOpen }()

	if err := waitForProcessExit(1234, time.Second); err == nil {
		t.Fatal("expected access denied to block replacement")
	}
}

func TestFormatWindowsCommandLine(t *testing.T) {
	line := formatWindowsCommandLine([]string{"C:\\Program Files\\cmdr\\cmdr.exe", "update-helper", "--target", "C:\\Program Files\\cmdr\\cmdr.exe"})
	if !strings.Contains(line, `"C:\Program Files\cmdr\cmdr.exe"`) {
		t.Fatalf("command line missing quoted executable: %s", line)
	}
	if !strings.Contains(line, "update-helper") {
		t.Fatalf("command line missing helper: %s", line)
	}
}
