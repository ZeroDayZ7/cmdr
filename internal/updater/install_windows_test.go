//go:build windows

package updater

import (
	"os"
	"path/filepath"
	"testing"
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
