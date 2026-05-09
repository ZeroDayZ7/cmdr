package general

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/utils"
)

var (
	output          string
	exclude         []string
	format          string
	copyToClipboard bool
)

func NewTreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tree [path]",
		Short: "Display a tree view of folders and files",
		Long: `Example:
  cmdr tree
  cmdr tree ./my-folder -o tree.txt
  cmdr tree ./my-folder -x .git,bin,node_modules
  cmdr tree --generate-ignore
  cmdr tree -f json
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			genIgnore, _ := cmd.Flags().GetBool("generate-ignore")
			if genIgnore {
				return utils.GenerateFileWithDefaults(".cmdrignore", nil)
			}

			path := "."
			if len(args) > 0 {
				path = args[0]
			}

			info, err := os.Stat(path)
			if err != nil {
				return err
			}

			defaultIgnore, _ := utils.ReadIgnoreFile(".cmdrignore")
			excludeList := append(defaultIgnore, utils.ParseCommaSeparated(strings.Join(exclude, ","))...)

			var result string

			switch format {
			case "ascii":
				var builder strings.Builder
				abs, _ := filepath.Abs(path)
				builder.WriteString(filepath.Base(abs) + "\n")
				if err := printTree(path, "", excludeList, &builder); err != nil {
					return err
				}
				result = builder.String()

			case "json":
				node, err := buildJSONTree(path, excludeList)
				if err != nil {
					return err
				}
				data, _ := json.MarshalIndent(node, "", "  ")
				result = string(data)

			case "csv":
				records := [][]string{{"path", "type"}}
				if err := buildCSVTree(path, "", excludeList, &records); err != nil {
					return err
				}
				var buf bytes.Buffer
				w := csv.NewWriter(&buf)
				if err := w.WriteAll(records); err != nil {
					return fmt.Errorf("failed to write CSV: %w", err)
				}
				result = buf.String()

			case "md":
				var builder strings.Builder
				builder.WriteString("# " + info.Name() + "\n")
				if err := printMarkdownTree(path, "", excludeList, &builder); err != nil {
					return err
				}
				result = builder.String()

			default:
				return fmt.Errorf("unknown format: %s", format)
			}

			if output != "" {
				if err := os.WriteFile(output, []byte(result), 0644); err != nil {
					return err
				}
				fmt.Printf("Tree saved to %s\n", output)
			} else if !copyToClipboard {
				fmt.Print(result)
			}
			if copyToClipboard {
				if err := utils.CopyToClipboard(result); err != nil {
					return err
				}
				fmt.Println("\n(Content copied to clipboard)")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file to save the tree")
	cmd.Flags().StringSliceVarP(&exclude, "exclude", "x", []string{}, "Comma-separated list of folders/files/extensions to exclude")
	cmd.Flags().BoolP("generate-ignore", "g", false, "Generate a default .cmdrignore file")
	cmd.Flags().StringVarP(&format, "format", "f", "ascii", "Output format: ascii|json|csv|md")
	cmd.Flags().BoolVarP(&copyToClipboard, "copy", "c", false, "Copy output to clipboard")

	return cmd
}

type Node struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Children []Node `json:"children,omitempty"`
}

func buildJSONTree(path string, excludeList []string) (Node, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Node{}, err
	}
	node := Node{Name: info.Name(), Type: "file"}
	if info.IsDir() {
		node.Type = "dir"
		entries, _ := os.ReadDir(path)
		for _, e := range entries {
			if utils.ShouldExclude(e.Name(), excludeList) {
				continue
			}
			child, _ := buildJSONTree(filepath.Join(path, e.Name()), excludeList)
			node.Children = append(node.Children, child)
		}
	}
	return node, nil
}

func buildCSVTree(path, prefix string, excludeList []string, records *[][]string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if utils.ShouldExclude(entry.Name(), excludeList) {
			continue
		}
		row := []string{filepath.Join(prefix, entry.Name()), "file"}
		if entry.IsDir() {
			row[1] = "dir"
		}
		*records = append(*records, row)
		if entry.IsDir() {
			if err := buildCSVTree(filepath.Join(path, entry.Name()), filepath.Join(prefix, entry.Name()), excludeList, records); err != nil {
				return err
			}
		}
	}
	return nil
}

func printMarkdownTree(path, prefix string, excludeList []string, builder *strings.Builder) error {
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
			if err := printMarkdownTree(filepath.Join(path, entry.Name()), prefix+"  ", excludeList, builder); err != nil {
				return err
			}
		}
	}
	return nil
}

func printTree(path, prefix string, excludeList []string, builder *strings.Builder) error {
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
		connector := "├── "
		if i == len(filtered)-1 {
			connector = "└── "
		}
		fmt.Fprintf(builder, "%s%s%s\n", prefix, connector, entry.Name())

		if entry.IsDir() {
			newPrefix := prefix
			if i == len(filtered)-1 {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			if err := printTree(filepath.Join(path, entry.Name()), newPrefix, excludeList, builder); err != nil {
				return err
			}
		}
	}
	return nil
}
