package updater

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExecutablePathRejectsDirectory(t *testing.T) {
	info, err := os.Stat(".")
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().IsRegular() {
		t.Fatal("test setup unexpectedly points to a regular file")
	}
}

func TestPrepareStagedBinary(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "source")
	if err := os.WriteFile(binary, []byte("new-binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	staged, err := PrepareStagedBinary(binary, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(staged)
	data, err := os.ReadFile(staged)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "new-binary" {
		t.Fatalf("staged content = %q", data)
	}
}

func TestPrepareStagedBinaryClearsOnMissingSource(t *testing.T) {
	dir := t.TempDir()
	if _, err := PrepareStagedBinary(filepath.Join(dir, "missing"), dir); err == nil {
		t.Fatal("expected missing source to fail")
	}
}

func TestUnixInstallationRollbackAndPermissions(t *testing.T) {
	if os.PathSeparator == '\\' {
		t.Skip("Unix-specific test")
	}
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

func TestValidateUpdatePaths(t *testing.T) {
	relative := filepath.Join("relative", "cmdr.exe")
	if err := ValidateUpdatePaths(relative, relative, relative); err == nil {
		t.Fatal("expected relative paths to fail")
	}
	absoluteDir := t.TempDir()
	if err := ValidateUpdatePaths(filepath.Join(absoluteDir, "cmdr.exe"), filepath.Join(absoluteDir, "staged.tmp"), filepath.Join(absoluteDir, "result")); err != nil {
		t.Fatalf("absolute paths rejected: %v", err)
	}
}

func TestValidateUpdatePathsRejectsEmpty(t *testing.T) {
	base := t.TempDir()
	if err := ValidateUpdatePaths("", filepath.Join(base, "staged"), filepath.Join(base, "result")); err == nil {
		t.Fatal("expected empty target to fail")
	}
}
