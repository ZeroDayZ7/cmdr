package cmd

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

// Remove-comments command
var removeCommentsCmd = &cobra.Command{
	Use:     "remove-comments",
	Aliases: []string{"rmc", "clean-comments"},
	Short:   "Remove comments from a specific file or all files in a directory",
	Long: `This command removes comments from the given file (any extension) or all files in a directory (any extension).
Example:
  cmdr remove-comments -f wikiData.tsx
  cmdr remove-comments -d ./my-folder`,
	Run: func(cmd *cobra.Command, args []string) {
		if directory != "" {
			err := removeCommentsFromDir(directory)
			if err != nil {
				fmt.Println("Error:", err)
			}
			return
		}
		if file != "" {
			err := removeCommentsFromFile(file)
			if err != nil {
				fmt.Println("Error:", err)
			}
			return
		}
		fmt.Println("Please provide a file (-f) or a directory (-d) to remove comments.")
	},
}

// Function to remove comments from a file
func removeCommentsFromFile(filePath string) error {
	// Open the file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	// Remove comments
	cleanedContent := removeComments(string(content))

	// Save the cleaned file
	err = os.WriteFile(filePath, []byte(cleanedContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", filePath, err)
	}

	fmt.Printf("Comments removed from the file: %s\n", filePath)
	return nil
}

// Function to remove comments from code
func removeComments(content string) string {
	// Remove multi-line comments
	multiLineRegex := regexp.MustCompile(`(?s)/\*.*?\*/`)
	cleaned := multiLineRegex.ReplaceAllString(content, "")

	// Remove single-line comments, preserving those in URLs (http://, https://)
	singleLineRegex := regexp.MustCompile(`(?m)(^|\s+)//.*$`)
	cleaned = singleLineRegex.ReplaceAllString(cleaned, "$1")

	return cleaned
}

// Function to remove comments from all files in a directory
func removeCommentsFromDir(directory string) error {
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, process only files
		if !info.IsDir() {
			err := removeCommentsFromFile(path)
			if err != nil {
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

func init() {
	rootCmd.AddCommand(removeCommentsCmd)
	removeCommentsCmd.Flags().StringVarP(&file, "file", "f", "", "Path to the file from which comments should be removed")
	removeCommentsCmd.Flags().StringVarP(&directory, "directory", "d", "", "Path to the folder from which comments should be removed")
}
