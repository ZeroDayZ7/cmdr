package annotate

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/zerodayz7/cmdr/internal/profiles"
)

func TestProcessBatch_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()
	cfg := Config{
		ProfilesConfig: &profiles.Config{
			CommentStyles: map[string]string{".go": "// %s"},
		},
	}

	// Tworzymy dużą liczbę plików
	var paths []string
	for i := 0; i < 100; i++ {
		p := filepath.Join(tempDir, "file_"+string(rune(i))+".go")
		os.WriteFile(p, []byte("package test"), 0644)
		paths = append(paths, p)
	}

	// Tworzymy kontekst, który od razu anulujemy
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results := ProcessBatch(ctx, paths, cfg)

	// Przy anulowanym kontekście nie powinniśmy przetworzyć wszystkich plików
	if len(results) == len(paths) {
		t.Errorf("Expected fewer results due to cancellation, got %d", len(results))
	}
}
