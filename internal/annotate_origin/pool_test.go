package annotate_origin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/zerodayz7/cmdr/internal/annotate"
	"github.com/zerodayz7/cmdr/internal/profiles"
)

func TestProcessBatch_ContextCancellation(t *testing.T) {
	tempDir := t.TempDir()

	cfg := annotate.Config{
		ProfilesConfig: &profiles.Config{
			CommentStyles: map[string]string{".go": "// %s"},
		},
	}

	var paths []string
	for i := range 100 {
		p := filepath.Join(tempDir, fmt.Sprintf("file_%d.go", i))

		if err := os.WriteFile(p, []byte("package test"), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", p, err)
		}
		paths = append(paths, p)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results := ProcessBatch(ctx, paths, cfg)

	if len(results) == len(paths) {
		t.Errorf("Expected fewer results due to cancellation, got %d", len(results))
	}
}
