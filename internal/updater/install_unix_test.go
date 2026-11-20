//go:build !windows

package updater

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnixInstallationFailureRestoresOriginal(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "cmdr")
	staged := filepath.Join(dir, "replacement")
	if err := os.WriteFile(current, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staged, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}

	oldRename := unixRename
	unixRename = func(from, to string) error {
		if from == staged {
			return os.ErrPermission
		}
		return oldRename(from, to)
	}
	defer func() { unixRename = oldRename }()

	if err := ApplyPlatformUpdate(current, staged); err == nil {
		t.Fatal("expected staged rename failure")
	}
	data, err := os.ReadFile(current)
	if err != nil || string(data) != "old" {
		t.Fatal("old executable was not restored")
	}
}

func TestUnixRollbackFailurePreservesBackup(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "cmdr")
	staged := filepath.Join(dir, "replacement")
	if err := os.WriteFile(current, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staged, []byte("new"), 0o755); err != nil {
		t.Fatal(err)
	}

	oldRename := unixRename
	stageRenameFailed := false
	unixRename = func(from, to string) error {
		if from == staged && !stageRenameFailed {
			stageRenameFailed = true
			return os.ErrPermission
		}
		if to == current && from != staged {
			return os.ErrPermission
		}
		return oldRename(from, to)
	}
	defer func() { unixRename = oldRename }()

	err := ApplyPlatformUpdate(current, staged)
	if err == nil || !strings.Contains(err.Error(), "rollback failed") {
		t.Fatalf("expected rollback failure, got %v", err)
	}
}

func TestUnixInstallationRollbackAndPermissions(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "cmdr")
	staged := filepath.Join(dir, "replacement")
	if err := os.WriteFile(current, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(staged, []byte("new"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := ApplyPlatformUpdate(current, staged); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(current)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new" {
		t.Fatalf("installed content = %q", data)
	}
	info, err := os.Stat(current)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		t.Fatal("installed binary is not executable")
	}
}
