package annotate

import (
	"testing"

	"github.com/zerodayz7/cmdr/internal/profiles"
)

func TestShouldIgnore(t *testing.T) {
	cfg := Config{
		ProfilesConfig: &profiles.Config{
			Global: profiles.GlobalConfig{Ignore: []string{".git", "vendor"}},
		},
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"src/main.go", false},
		{".git/config", true},
		{"internal/vendor/lib.go", true},
		{"vendor/main.go", true},
		{"my_vendor_stuff/file.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := shouldIgnore(tt.path, cfg); got != tt.expected {
				t.Errorf("shouldIgnore(%s) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
