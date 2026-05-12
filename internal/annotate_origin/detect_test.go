package annotate_origin

import "testing"

func TestHasAnnotation(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{"Go comment", "// cmdr: main.go\npackage main", true},
		{"Python comment", "# cmdr: script.py\nimport os", true},
		{"SQL comment", "-- cmdr: schema.sql\nCREATE TABLE", true},
		{"No annotation", "package main\nfunc main() {}", false},
		{"Too deep", "\n\n\n\n\n\n\n\n\n\n\n// cmdr: hidden", false},
		{"Trim space", "   // cmdr: spaced", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HasAnnotation(tt.content); got != tt.expected {
				t.Errorf("%s: got %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}
