package utils

import (
	"path/filepath"
	"strings"
)

func ParseCommaSeparated(input string) []string {
	var items []string
	for _, s := range strings.Split(input, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			items = append(items, s)
		}
	}
	return items
}

func ShouldExclude(name string, excludeList []string) bool {
	name = strings.TrimSpace(name)
	if name == "" {
		return false
	}

	for _, e := range excludeList {
		e = strings.TrimSpace(e)
		if e == "" || strings.HasPrefix(e, "#") {
			continue
		}

		if name == e {
			return true
		}

		cleanExt := strings.TrimPrefix(e, "*")
		if strings.HasPrefix(cleanExt, ".") && strings.HasSuffix(name, cleanExt) {
			return true
		}

		if strings.HasSuffix(e, "/") || strings.HasSuffix(e, string(filepath.Separator)) {
			cleanFolder := strings.TrimRight(e, "/"+string(filepath.Separator))
			if name == cleanFolder {
				return true
			}
		}
	}

	return false
}
