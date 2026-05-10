package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/utils"
)

var (
	regDir  string
	regFile string
)

// #region NewAnnotateRegionsCmd
func NewAnnotateRegionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "annotate-regions",
		Aliases: []string{"reg", "ann", "regions"},
		Short:   "Wraps every function in a file with #region and #endregion",
		Example: "cmdr reg -d ./internal/api\ncmdr reg -f main.go",
		Run: func(cmd *cobra.Command, args []string) {
			if regFile != "" {
				if err := annotateFile(regFile); err != nil {
					fmt.Printf("Error processing file %s: %v\n", regFile, err)
				}
				return
			}

			if regDir == "" {
				regDir = "."
			}
			if err := annotateDir(regDir); err != nil {
				fmt.Printf("Error processing directory %s: %v\n", regDir, err)
			}
		},
	}

	cmd.Flags().StringVarP(&regDir, "directory", "d", "", "Directory to process")
	cmd.Flags().StringVarP(&regFile, "file", "f", "", "Specific file to process")

	return cmd
}

// #region annotateDir
func annotateDir(directory string) error {
	ignoreList, err := utils.ReadIgnoreFile(".cmdr_reg_ignore")
	if err != nil {
		return err
	}

	if ignoreList == nil {
		utils.GenerateFileWithDefaults(".cmdr_reg_ignore", []string{".git", "node_modules", "bin", ".g.dart", "vendor"})
		ignoreList, _ = utils.ReadIgnoreFile(".cmdr_reg_ignore")
	}

	return filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if utils.ShouldExclude(info.Name(), ignoreList) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			return annotateFile(path)
		}

		return nil
	})
}

// #region annotateFile
func annotateFile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	strContent := string(content)

	re := regexp.MustCompile(`(?m)^func\s+(?:\([^\)]+\)\s+)?([a-zA-Z0-9_]+)\s*\(`)

	newContent := re.ReplaceAllStringFunc(strContent, func(match string) string {

		submatch := re.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}
		funcName := submatch[1]

		regionTag := fmt.Sprintf("// #region %s", funcName)

		if strings.Contains(strContent, regionTag) {
			return match
		}

		return fmt.Sprintf("%s\n%s", regionTag, match)
	})

	if newContent != strContent {
		err = os.WriteFile(filePath, []byte(newContent), 0644)
		if err == nil {
			fmt.Printf("Annotated: %s\n", filePath)
		}
	}

	return err
}
