package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
)

var (
	file      string
	directory string
)

// #region RemoveCommentsCmd
// Removes comments from a file or all files in a directory.
func NewRemoveCommentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-comments",
		Aliases: []string{"rmc", "clean-comments"},
		Short:   "Remove comments from a specific file or all files in a directory",
		Long: `This command removes comments from the given file (any extension) or all files in a directory (any extension).
Example:
  cmdr remove-comments -f wikiData.tsx
  cmdr remove-comments -d ./my-folder`,
		Run: func(cmd *cobra.Command, args []string) {
			if directory != "" {
				if err := removeCommentsFromDir(directory); err != nil {
					fmt.Println("Error:", err)
				}
				return
			}
			if file != "" {
				if err := removeCommentsFromFile(file); err != nil {
					fmt.Println("Error:", err)
				}
				return
			}
			fmt.Println("Please provide a file (-f) or a directory (-d) to remove comments.")
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Path to the file from which comments should be removed")
	cmd.Flags().StringVarP(&directory, "directory", "d", "", "Path to the folder from which comments should be removed")
	return cmd
}

// #region removeCommentsFromFile
func removeCommentsFromFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	cleanedContent := removeComments(string(content))

	if err := os.WriteFile(filePath, []byte(cleanedContent), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %v", filePath, err)
	}

	fmt.Printf("Comments removed from the file: %s\n", filePath)
	return nil
}

// #region removeComments
func removeComments(content string) string {
	// Remove multi-line comments
	multiLineRegex := regexp.MustCompile(`(?s)/\*.*?\*/`)
	cleaned := multiLineRegex.ReplaceAllString(content, "")

	// Remove single-line comments, preserving those in URLs
	singleLineRegex := regexp.MustCompile(`(?m)(^|\s+)//.*$`)
	cleaned = singleLineRegex.ReplaceAllString(cleaned, "$1")

	return cleaned
}

// #region removeCommentsFromDir
func removeCommentsFromDir(directory string) error {
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if err := removeCommentsFromFile(path); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to process directory %s: %v", directory, err)
	}

	fmt.Printf("Comments removed from all files in the directory: %s\n", directory)
	return nil
}
