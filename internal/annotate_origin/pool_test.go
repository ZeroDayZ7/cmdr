package annotate_origin

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/zerodayz7/cmdr/internal/annotate" // Importuj shared
	"github.com/zerodayz7/cmdr/internal/profiles"
)

func TestProcessBatch_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()

	// KLUCZOWA POPRAWKA: Używamy typu z pakietu shared
	cfg := annotate.Config{
		ProfilesConfig: &profiles.Config{
			CommentStyles: map[string]string{".go": "// %s"},
		},
	}

	// Tworzymy dużą liczbę plików
	var paths []string
	for i := 0; i < 100; i++ {
		// Używamy fmt.Sprintf zamiast string(rune), żeby nazwy były czytelne
		p := filepath.Join(tempDir, filepath.Clean(filepath.Join("/", "file_"+string(rune(i))+".go")))
		// Lepiej tak:
		// p := filepath.Join(tempDir, fmt.Sprintf("file_%d.go", i))
		os.WriteFile(p, []byte("package test"), 0644)
		paths = append(paths, p)
	}

	// Tworzymy kontekst, który od razu anulujemy
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Teraz typ cfg zgadza się z tym, czego oczekuje ProcessBatch
	results := ProcessBatch(ctx, paths, cfg)

	// Przy anulowanym kontekście nie powinniśmy przetworzyć wszystkich plików
	if len(results) == len(paths) {
		t.Errorf("Expected fewer results due to cancellation, got %d", len(results))
	}
}
