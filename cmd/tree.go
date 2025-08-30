package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/username/cli/internal/utils"
)

var (
	output  string
	exclude []string
)

var treeCmd = &cobra.Command{
	Use:   "tree [path]",
	Short: "Display a tree view of folders and files",
	Long: `Example:
  cmdr tree
  cmdr tree ./my-folder -o tree.txt
  cmdr tree ./my-folder -x .git,bin,node_modules
  cmdr tree --generate-ignore
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

		var builder strings.Builder
		err := printTree(path, "", excludeList, &builder)
		if err != nil {
			return err
		}

		if output != "" {
			return os.WriteFile(output, []byte(builder.String()), 0644)
		}

		fmt.Print(builder.String())
		return nil
	},
}

func init() {
	treeCmd.Flags().StringVarP(&output, "output", "o", "", "Output file to save the tree")
	treeCmd.Flags().StringSliceVarP(&exclude, "exclude", "x", []string{}, "Comma-separated list of folders/files/extensions to exclude")
	treeCmd.Flags().BoolP("generate-ignore", "g", false, "Generate a default .cmdrignore file")
	rootCmd.AddCommand(treeCmd)
}

// generateIgnoreFile tworzy plik .cmdrignore w katalogu z exe
func generateIgnoreFile() error {
	defaultList := []string{".git", "node_modules", "bin", "dist", ".DS_Store"}
	content := strings.Join(defaultList, "\n")

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)

	ignorePath := filepath.Join(exeDir, ".cmdrignore")
	err = os.WriteFile(ignorePath, []byte(content), 0644)
	if err == nil {
		fmt.Printf("Generated .cmdrignore at %s with default values.\n", ignorePath)
	}
	return err
}

// readDefaultIgnore wczytuje plik .cmdrignore z katalogu z exe
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

// printTree generuje drzewo katalogów w formacie ASCII
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
		line := fmt.Sprintf("%s%s%s\n", prefix, connector, entry.Name())
		builder.WriteString(line)

		if entry.IsDir() {
			newPrefix := prefix
			if i == len(entries)-1 {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}
			err := printTree(filepath.Join(path, entry.Name()), newPrefix, excludeList, builder)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
