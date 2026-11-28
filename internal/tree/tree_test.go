package tree

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateRejectsFilePath(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "sample.txt")
	if err := os.WriteFile(file, []byte("test"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Generate(TreeOptions{Path: file, Format: "ascii", MaxDepth: -1})
	if err == nil {
		t.Fatal("expected file path to be rejected")
	}
	if !strings.Contains(err.Error(), "must be a directory") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateRepeatedDirectoryState(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	result, err := Generate(TreeOptions{Path: dir, Format: "ascii", MaxDepth: -1})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "a.txt") {
		t.Fatalf("missing file in tree: %s", result)
	}
}
