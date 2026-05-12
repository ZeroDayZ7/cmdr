package annotate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zerodayz7/cmdr/internal/profiles"
)

func TestIsBinary(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected bool
	}{
		{"Plain text", []byte("package main\n\nfunc main() {}"), false},
		{"Empty file", []byte(""), false},
		{"Binary PNG", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00}, true},
		{"Text with special chars", []byte("Hello Gopher! 🚀"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBinary(tt.input); got != tt.expected {
				t.Errorf("IsBinary() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// annotate_test.go (dopisz pod TestIsBinary)

func TestInjectComment(t *testing.T) {
	comment := "// cmdr: test/path.go"

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			"Standard file",
			"package main",
			comment + "\n\npackage main",
		},
		{
			"File with shebang",
			"#!/bin/bash\necho 1",
			"#!/bin/bash\n" + comment + "\necho 1",
		},
		{
			"Shebang only",
			"#!/usr/bin/env node",
			"#!/usr/bin/env node\n" + comment,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := injectComment(tt.content, comment)
			if got != tt.expected {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.expected)
			}
		})
	}
}

func TestIsAllowedExtension(t *testing.T) {
	cfg := Config{
		ProfilesConfig: &profiles.Config{
			CommentStyles: map[string]string{".go": "// %s", ".py": "# %s"},
		},
	}

	t.Run("Global allow", func(t *testing.T) {
		if !isAllowedExtension(".go", cfg) {
			t.Error("Expected .go to be allowed")
		}
	})

	t.Run("Profile override", func(t *testing.T) {
		cfg.Profile = &profiles.Profile{Extensions: []string{".dart"}}
		if isAllowedExtension(".go", cfg) {
			t.Error("Expected .go to be blocked when profile limits to .dart")
		}
		if !isAllowedExtension(".dart", cfg) {
			t.Error("Expected .dart to be allowed by profile")
		}
	})
}

func TestAtomicWrite(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "atomic.txt")
	data := []byte("hello atomic")

	// Test podstawowego zapisu
	if err := atomicWrite(filePath, data); err != nil {
		t.Fatalf("atomicWrite failed: %v", err)
	}

	got, _ := os.ReadFile(filePath)
	if string(got) != string(data) {
		t.Errorf("got %s, want %s", string(got), string(data))
	}

	// Sprawdzenie czy nie ma plików .tmp
	files, _ := os.ReadDir(tempDir)
	for _, f := range files {
		if strings.Contains(f.Name(), ".tmp") {
			t.Errorf("Temporary file leaked: %s", f.Name())
		}
	}
}

func TestAnnotateFile_Integration(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "main.go")
	os.WriteFile(path, []byte("package main"), 0644)

	cfg := Config{
		ProfilesConfig: &profiles.Config{
			CommentStyles: map[string]string{".go": "// %s"},
		},
	}

	err := AnnotateFile(path, cfg)
	if err != nil {
		t.Fatalf("AnnotateFile failed: %v", err)
	}

	content, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(content), "//") {
		t.Error("File was not annotated")
	}
	if !strings.Contains(string(content), "main.go") {
		t.Error("Annotation does not contain filename")
	}
}
