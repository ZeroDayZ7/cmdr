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

		// Dopasowanie dokładnej nazwy pliku/folderu
		if name == e {
			return true
		}

		// Dopasowanie po rozszerzeniu (*.md)
		if strings.HasPrefix(e, "*.") && strings.HasSuffix(name, e[1:]) {
			return true
		}

		// Dopasowanie folderów względnych (np. ./bin)
		if strings.HasSuffix(e, string(filepath.Separator)) && name == strings.TrimSuffix(e, string(filepath.Separator)) {
			return true
		}
	}

	return false
}
