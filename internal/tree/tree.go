package tree

import (
	"fmt"
	"os"
	"strings"
)

func Generate(opts TreeOptions) (string, error) {
	info, err := os.Stat(opts.Path)
	if err != nil {
		return "", fmt.Errorf("cannot access tree path %q: %w", opts.Path, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("tree path must be a directory: %s", opts.Path)
	}

	switch strings.ToLower(opts.Format) {
	case "ascii":
		return formatASCII(opts.Path, opts.ExcludeList, opts.MaxDepth)
	case "json":
		return formatJSON(opts.Path, opts.ExcludeList, opts.MaxDepth)
	case "csv":
		return formatCSV(opts.Path, opts.ExcludeList, opts.MaxDepth)
	case "md":
		return formatMarkdown(opts.Path, opts.ExcludeList, opts.MaxDepth)
	default:
		return "", fmt.Errorf("unsupported format: %s", opts.Format)
	}
}
