package imageutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceFileKeepsExistingFileWhenOverwriteDisabled(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "new.txt")
	dst := filepath.Join(dir, "existing.txt")
	if err := os.WriteFile(src, []byte("new"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(dst, []byte("old"), 0o644); err != nil {
		t.Fatalf("write destination: %v", err)
	}

	if err := ReplaceFile(src, dst, false); err == nil {
		t.Fatal("expected overwrite=false to reject replacing an existing file")
	}
	content, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read destination after rejected replace: %v", err)
	}
	if string(content) != "old" {
		t.Fatalf("destination content changed unexpectedly: got %q want %q", string(content), "old")
	}
}

func TestReplaceFileOverwritesExistingFileWhenAllowed(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "new.txt")
	dst := filepath.Join(dir, "existing.txt")
	if err := os.WriteFile(src, []byte("new"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(dst, []byte("old"), 0o644); err != nil {
		t.Fatalf("write destination: %v", err)
	}

	if err := ReplaceFile(src, dst, true); err != nil {
		t.Fatalf("ReplaceFile returned error for overwrite=true: %v", err)
	}
	content, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read destination after overwrite: %v", err)
	}
	if string(content) != "new" {
		t.Fatalf("destination content was not replaced: got %q want %q", string(content), "new")
	}
}

func TestNormalizeOutputFormatRejectsInvalidExtension(t *testing.T) {
	if _, err := NormalizeOutputFormat("image.bin", "png"); err == nil {
		t.Fatal("expected invalid extension to be rejected")
	}
}

func TestNormalizeOutputFormatRequiresMatchingExtensionAndFormat(t *testing.T) {
	if _, err := NormalizeOutputFormat("image.png", "jpeg"); err == nil {
		t.Fatal("expected PNG path to reject JPEG format")
	}
}
