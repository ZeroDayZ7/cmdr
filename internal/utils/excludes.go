package utils

import (
	"path/filepath"
	"strings"
)

// #region ParseCommaSeparated
// Splits a comma-separated string into trimmed, non-empty items.
func ParseCommaSeparated(input string) []string {
	var items []string
	for _, s := range strings.Split(input, ",") { // używamy Split zamiast SplitSeq
		s = strings.TrimSpace(s)
		if s != "" {
			items = append(items, s)
		}
	}
	return items
}

// #region ShouldExclude
// Checks if a file or folder should be excluded based on names, extensions, or folder paths.
func ShouldExclude(name string, excludeList []string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}

	for _, e := range excludeList {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}

		// Exact match for file/folder names
		if name == e {
			return true
		}

		// Match by extension (*.ext)
		if strings.HasPrefix(e, "*.") && strings.HasSuffix(name, e[1:]) {
			return true
		}

		// Match relative folders (e.g., ./bin)
		if strings.HasSuffix(e, string(filepath.Separator)) &&
			name == strings.TrimSuffix(e, string(filepath.Separator)) {
			return true
		}
	}

	return false
}
