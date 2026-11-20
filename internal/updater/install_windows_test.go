//go:build windows

package updater

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteUpdateResultAtomic(t *testing.T) {
	dir := t.TempDir()
	result := filepath.Join(dir, "update.result")
	writeUpdateResult(result, "replace failed")
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
