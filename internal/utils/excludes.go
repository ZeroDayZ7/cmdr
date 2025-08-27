package utils

import (
	"path/filepath"
	"strings"
)

// ParseCommaSeparated splits string by comma and trims spaces
func ParseCommaSeparated(input string) []string {
	var items []string
	for s := range strings.SplitSeq(input, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			items = append(items, s)
		}
	}
	return items
}

// ShouldExclude checks if a file or folder should be excluded
// Supports folder names, file names, and extensions (e.g., ".exe")
func ShouldExclude(path string, excludeList []string) bool {
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(path))

	for _, ex := range excludeList {
		ex = strings.ToLower(strings.TrimSpace(ex))
		if ex == "" {
			continue
		}

		if ex[0] == '.' {
			// treat as extension
			if ext == ex {
				return true
			}
		} else {
			// treat as folder/file name
			if base == ex || strings.Contains(path, string(filepath.Separator)+ex) {
				return true
			}
		}
	}
	return false
}
