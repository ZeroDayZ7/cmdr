package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/utils"
)

var (
	file      string
	directory string
)

// #region RemoveCommentsCmd
func NewRemoveCommentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-comments",
		Aliases: []string{"rmc", "clean-comments"},
		Short:   "Remove comments from a specific file or all files in a directory",
		Long: `This command removes comments from the given file or all files in a directory.
It uses a dedicated ignore file: .cmdr_rmc_ignore located in the executable directory.`,
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
	ignoreFileName := ".cmdr_rmc_ignore"

	// 1. Próbujemy wczytać plik
	ignoreList, err := utils.ReadIgnoreFile(ignoreFileName)

	// 2. Jeśli plik nie istnieje (ignoreList jest nil), generujemy go automatycznie
	if ignoreList == nil {
		// Definiujemy sensowne domyślne wzorce dla usuwania komentarzy
		defaults := []string{
			".g.dart",
			".git",
			"node_modules",
			"bin",
			"vendor",
		}

		err = utils.GenerateFileWithDefaults(ignoreFileName, defaults)
		if err != nil {
			fmt.Printf("Warning: Could not create default ignore file: %v\n", err)
		} else {
			// Po wygenerowaniu, wczytujemy go ponownie, aby mieć listę
			ignoreList, _ = utils.ReadIgnoreFile(ignoreFileName)
		}
	}

	// 3. Dalej idzie standardowy Walk
	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Sprawdzanie wykluczeń z użyciem Twojego utils.ShouldExclude
		if utils.ShouldExclude(info.Name(), ignoreList) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() {
			if err := removeCommentsFromFile(path); err != nil {
				return err
			}
		}
		return nil
	})

	return err
}
