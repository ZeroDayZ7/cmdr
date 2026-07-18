package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zerodayz7/cmdr/internal/utils"
)

func formatASCII(path string, excludeList []string, maxDepth int) (string, error) {
	var builder strings.Builder
	abs, _ := filepath.Abs(path)

	// Rozbite wywołania eliminują niepotrzebną alokację pamięci
	builder.WriteString(filepath.Base(abs))
	builder.WriteString("\n")

	err := walkASCII(path, "", excludeList, &builder, 0, maxDepth)
	return builder.String(), err
}

func walkASCII(path, prefix string, excludeList []string, builder *strings.Builder, currentDepth, maxDepth int) error {
	if maxDepth >= 0 && currentDepth > maxDepth {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	var filtered []os.DirEntry
	for _, entry := range entries {
		if !utils.ShouldExclude(entry.Name(), excludeList) {
			filtered = append(filtered, entry)
		}
	}

	for i, entry := range filtered {
		isLast := i == len(filtered)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		fmt.Fprintf(builder, "%s%s%s\n", prefix, connector, entry.Name())

		if entry.IsDir() {
			newPrefix := prefix
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			err := walkASCII(filepath.Join(path, entry.Name()), newPrefix, excludeList, builder, currentDepth+1, maxDepth)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
