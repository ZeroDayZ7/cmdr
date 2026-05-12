package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/profiles"
)

var (
	removeFile string
	removeDir  string
)

// #region NewRemoveCommentsCmd
func NewRemoveCommentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove-comments",
		Aliases: []string{"rmc", "clean-comments"},
		Short:   "Remove comments based on project profile",
		RunE:    runRemoveComments,
	}

	cmd.Flags().StringVarP(&removeFile, "file", "f", "", "Path to specific file")
	cmd.Flags().StringVarP(&removeDir, "directory", "d", "", "Path to directory")
	return cmd
}

// #region runRemoveComments
func runRemoveComments(cmd *cobra.Command, args []string) error {
	if removeFile != "" {
		return removeCommentsFromFile(removeFile)
	}

	targetDir := removeDir
	if targetDir == "" {
		targetDir = "."
	}

	return removeCommentsFromDir(targetDir)
}

// #region removeCommentsFromFile
func removeCommentsFromFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	cleanedContent := removeComments(string(content))

	if err := os.WriteFile(filePath, []byte(cleanedContent), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	fmt.Printf("✨ Cleaned: %s\n", filePath)
	return nil
}

// #region removeComments
func removeComments(content string) string {

	multiLineRegex := regexp.MustCompile(`(?s)/\*.*?\*/`)
	cleaned := multiLineRegex.ReplaceAllString(content, "")

	singleLineRegex := regexp.MustCompile(`(?m)(^|\s+)//.*$`)
	cleaned = singleLineRegex.ReplaceAllString(cleaned, "$1")

	return cleaned
}

// #region removeCommentsFromDir
func removeCommentsFromDir(directory string) error {

	cfg, err := profiles.LoadConfig()
	if err != nil {
		return err
	}

	activeProfile := profiles.DetectProjectProfile(directory, cfg)

	ignoredItems := cfg.Global.Ignore
	if activeProfile != nil {
		ignoredItems = append(ignoredItems, activeProfile.Ignore...)
		fmt.Printf("Using profile: %s for comment removal\n", activeProfile.Name)
	}

	return filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		name := info.Name()

		for _, ignored := range ignoredItems {
			if name == ignored {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		if !info.IsDir() {

			ext := filepath.Ext(name)
			if isSupportedExtension(ext) {
				if err := removeCommentsFromFile(path); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not clean %s: %v\n", path, err)
				}
			}
		}
		return nil
	})
}

// #region isSupportedExtension
func isSupportedExtension(ext string) bool {
	supported := []string{".go", ".dart", ".ts", ".js", ".tsx", ".jsx", ".c", ".cpp", ".java"}
	return slices.Contains(supported, ext)
}
