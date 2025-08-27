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
  cmdr tree ./my-folder -x .git,bin,node_modules`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		excludeList := utils.ParseCommaSeparated(strings.Join(exclude, ","))

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
	rootCmd.AddCommand(treeCmd)
}

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
