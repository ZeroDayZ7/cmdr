package annotate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildRelativePath(t *testing.T) {
	root := "/home/user/project"
	file := "/home/user/project/src/main.go"

	expected := "project/src/main.go"

	got, err := BuildRelativePath(root, file)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != expected {
		t.Errorf("got %s, want %s", got, expected)
	}
}

func TestFindProjectRoot_Monorepo(t *testing.T) {
	// W testach w Go często tworzy się tymczasową strukturę katalogów
	tempDir := t.TempDir()

	repoRoot := filepath.Join(tempDir, "repo")
	subApp := filepath.Join(repoRoot, "apps/api")

	os.MkdirAll(subApp, 0755)

	// Tworzymy go.mod w root
	os.WriteFile(filepath.Join(repoRoot, "go.mod"), []byte("module test"), 0644)
	// Tworzymy package.json w subApp
	os.WriteFile(filepath.Join(subApp, "package.json"), []byte("{}"), 0644)

	t.Run("Should find nearest root (subApp)", func(t *testing.T) {
		root, err := FindProjectRoot(subApp, Config{})
		if err != nil {
			t.Fatal(err)
		}
		if root != subApp {
			t.Errorf("Expected root to be %s, got %s", subApp, root)
		}
	})
}
