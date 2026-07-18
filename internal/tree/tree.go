package tree

import (
	"fmt"
	"strings"
)

func Generate(opts TreeOptions) (string, error) {
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
