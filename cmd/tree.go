package cmd

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
	output  string   // Output file path
	exclude []string // Files/folders to exclude
	format  string   // Output format: ascii|json|csv|md
)

var treeCmd = &cobra.Command{
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
			return generateIgnoreFile()
		}

		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		defaultIgnore, _ := readDefaultIgnore()
		excludeList := append(defaultIgnore, utils.ParseCommaSeparated(strings.Join(exclude, ","))...)

		switch format {
		case "ascii":
			var builder strings.Builder
			if err := printTree(path, "", excludeList, &builder); err != nil {
				return err
			}
			if output != "" {
				return os.WriteFile(output, []byte(builder.String()), 0644)
			}
			fmt.Print(builder.String())

		case "json":
			node, err := buildJSONTree(path, excludeList)
			if err != nil {
				return err
			}
			data, _ := json.MarshalIndent(node, "", "  ")
			if output != "" {
				return os.WriteFile(output, data, 0644)
			}
			fmt.Println(string(data))

		case "csv":
			records := [][]string{{"path", "type"}}
			if err := buildCSVTree(path, "", excludeList, &records); err != nil {
				return err
			}
			var buf bytes.Buffer
			w := csv.NewWriter(&buf)
			w.WriteAll(records)
			w.Flush()
			if output != "" {
				return os.WriteFile(output, buf.Bytes(), 0644)
			}
			fmt.Print(buf.String())

		case "md":
			var builder strings.Builder
			if err := printMarkdownTree(path, "", excludeList, &builder); err != nil {
				return err
			}
			if output != "" {
				return os.WriteFile(output, []byte(builder.String()), 0644)
			}
			fmt.Print(builder.String())

		default:
			return fmt.Errorf("unknown format: %s", format)
		}

		return nil
	},
}

func init() {
	treeCmd.Flags().StringVarP(&output, "output", "o", "", "Output file to save the tree")
	treeCmd.Flags().StringSliceVarP(&exclude, "exclude", "x", []string{}, "Comma-separated list of folders/files/extensions to exclude")
	treeCmd.Flags().BoolP("generate-ignore", "g", false, "Generate a default .cmdrignore file")
	treeCmd.Flags().StringVarP(&format, "format", "f", "ascii", "Output format: ascii|json|csv|md")
	rootCmd.AddCommand(treeCmd)
}

// Node represents a file/directory for JSON
type Node struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // "file" or "dir"
	Children []Node `json:"children,omitempty"`
}

// buildJSONTree builds a JSON tree structure
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

// buildCSVTree generates CSV records recursively
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
			buildCSVTree(filepath.Join(path, entry.Name()), filepath.Join(prefix, entry.Name()), excludeList, records)
		}
	}
	return nil
}

// printMarkdownTree generates Markdown formatted tree
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
			printMarkdownTree(filepath.Join(path, entry.Name()), prefix+"  ", excludeList, builder)
		}
	}
	return nil
}

// generateIgnoreFile creates a default .cmdrignore in executable folder
func generateIgnoreFile() error {
	defaultList := []string{".git", "node_modules", "bin", "dist", ".DS_Store"}
	content := strings.Join(defaultList, "\n")

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)

	ignorePath := filepath.Join(exeDir, ".cmdrignore")
	if err := os.WriteFile(ignorePath, []byte(content), 0644); err == nil {
		fmt.Printf("Generated .cmdrignore at %s with default values.\n", ignorePath)
	}
	return err
}

// readDefaultIgnore reads the .cmdrignore from executable folder
func readDefaultIgnore() ([]string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)

	ignorePath := filepath.Join(exeDir, ".cmdrignore")
	if _, err := os.Stat(ignorePath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(ignorePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var result []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			result = append(result, l)
		}
	}
	return result, nil
}

// printTree generates ASCII tree using fmt.Fprintf instead of WriteString(fmt.Sprintf(...))
func printTree(path, prefix string, excludeList []string, builder *strings.Builder) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for i, entry := range entries {
		if utils.ShouldExclude(entry.Name(), excludeList) {
			continue
		}
		connector := "├── "
		if i == len(entries)-1 {
			connector = "└── "
		}
		fmt.Fprintf(builder, "%s%s%s\n", prefix, connector, entry.Name())

		if entry.IsDir() {
			newPrefix := prefix
			if i == len(entries)-1 {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			printTree(filepath.Join(path, entry.Name()), newPrefix, excludeList, builder)
		}
	}
	return nil
}
