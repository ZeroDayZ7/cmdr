package tree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zerodayz7/cmdr/internal/utils"
)

func formatMarkdown(path string, excludeList []string, maxDepth int) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	fmt.Fprintf(&builder, "# %s\n\n", info.Name())

	err = walkMarkdown(path, "", excludeList, &builder, 0, maxDepth)
	return builder.String(), err
}

func walkMarkdown(path, prefix string, excludeList []string, builder *strings.Builder, currentDepth, maxDepth int) error {
	if maxDepth >= 0 && currentDepth > maxDepth {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if utils.ShouldExclude(entry.Name(), excludeList) {
			continue
		}

		fmt.Fprintf(builder, "%s- %s\n", prefix, entry.Name())

		if entry.IsDir() {
			err := walkMarkdown(filepath.Join(path, entry.Name()), prefix+"  ", excludeList, builder, currentDepth+1, maxDepth)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
