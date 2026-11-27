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
	defer func() { _ = os.Remove(staged) }()
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
