package annotate_origin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zerodayz7/cmdr/internal/annotate"
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
			if got := annotate.IsBinary(tt.input); got != tt.expected {
				t.Errorf("IsBinary() = %v, want %v", got, tt.expected)
			}
		})
	}
}

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
			got := annotate.InjectComment(tt.content, comment)
			if got != tt.expected {
				t.Errorf("got:\n%s\nwant:\n%s", got, tt.expected)
			}
		})
	}
}

func TestIsAllowedExtension(t *testing.T) {
	cfg := annotate.Config{
		ProfilesConfig: &profiles.Config{
			CommentStyles: map[string]string{".go": "// %s", ".py": "# %s"},
		},
	}

	t.Run("Global allow", func(t *testing.T) {
		if !annotate.IsAllowedExtension(".go", cfg) {
			t.Error("Expected .go to be allowed")
		}
	})

	t.Run("Profile override", func(t *testing.T) {
		cfg.Profile = &profiles.Profile{Extensions: []string{".dart"}}
		if annotate.IsAllowedExtension(".go", cfg) {
			t.Error("Expected .go to be blocked when profile limits to .dart")
		}
		if !annotate.IsAllowedExtension(".dart", cfg) {
			t.Error("Expected .dart to be allowed by profile")
		}
	})
}

func TestAtomicWrite(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "atomic.txt")
	data := []byte("hello atomic")

	if err := annotate.AtomicWrite(filePath, data); err != nil {
		t.Fatalf("atomicWrite failed: %v", err)
	}

	got, _ := os.ReadFile(filePath)
	if string(got) != string(data) {
		t.Errorf("got %s, want %s", string(got), string(data))
	}

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

	// 1. Sprawdzamy błąd zapisu - jeśli setup zawiedzie, nie ma sensu iść dalej
	err := os.WriteFile(path, []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("failed to setup test file: %v", err)
	}

	cfg := annotate.Config{
		ProfilesConfig: &profiles.Config{
			CommentStyles: map[string]string{".go": "// %s"},
		},
	}

	// 2. Wywołanie funkcji testowanej - err już jest zadeklarowany wyżej, więc używamy '='
	err = annotate.AnnotateFile(path, cfg, func(relPath string, style string) string {
		return "cmdr: " + relPath
	})
	if err != nil {
		t.Fatalf("AnnotateFile failed: %v", err)
	}

	// 3. Sprawdzamy błąd odczytu - ignorowanie go ('_') to ryzykowna praktyka
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read annotated file: %v", err)
	}

	// 4. Asercje
	result := string(content)
	if !strings.Contains(result, "cmdr:") {
		t.Error("File was not annotated: missing 'cmdr:' prefix")
	}
	if !strings.Contains(result, "main.go") {
		t.Error("Annotation does not contain filename 'main.go'")
	}
}
