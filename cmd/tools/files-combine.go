package tools

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/utils"
)

var (
	outputFile     string
	extensions     string
	excludeDirs    string
	generateIgnore bool
)

func NewFilesCombineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "files-combine",
		Aliases: []string{"fc", "combine-files"},
		Short:   "Combine contents of files with chosen extensions into one file",
		Long: `Example:
  cmdr files-combine
  cmdr files-combine -n combined.txt -e ts,tsx,json
  cmdr files-combine -x .git,node_modules
  cmdr files-combine --generate-ignore
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Generate ignore file if requested
			if generateIgnore {
				return utils.GenerateFileWithDefaults(".files-combineignore", []string{
					"node_modules",
					"dist",
					".git",
				})
			}

			if outputFile == "" {
				outputFile = "combined.txt"
			}
			if extensions == "" {
				extensions = "ts,tsx,jsx,json"
			}

			// Read ignore file + CLI excludes
			defaultIgnore, _ := utils.ReadIgnoreFile(".files-combineignore")
			excludeList := append(defaultIgnore, utils.ParseCommaSeparated(excludeDirs)...)

			extList := parseExtensions(extensions)

			file, err := os.Create(outputFile)
			if err != nil {
				return fmt.Errorf("error creating output file: %w", err)
			}
			defer file.Close()

			writer := bufio.NewWriter(file)

			// Initial empty lines
			for range [4]struct{}{} {
				writer.WriteString("\n")
			}

			err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() && utils.ShouldExclude(info.Name(), excludeList) {
					return filepath.SkipDir
				}

				if !info.IsDir() && hasAllowedExtension(path, extList) {
					_, fileName := filepath.Split(path)
					fmt.Fprintf(writer, "───────────── %s ─────────────\n", fileName)

					content, readErr := os.ReadFile(path)
					if readErr != nil {
						fmt.Fprintf(writer, "Error reading file: %v\n", readErr)
					} else {
						writer.Write(content)
					}
					writer.WriteString("\n\n")
				}
				return nil
			})

			if err != nil {
				return fmt.Errorf("error while scanning files: %w", err)
			}

			writer.Flush()
			fmt.Printf("✅ Finished! Combined file: %s\n", outputFile)
			return nil
		},
	}

	// Flags
	cmd.Flags().StringVarP(&extensions, "extensions", "e", "", "Comma-separated list of file extensions (e.g. ts,tsx,json)")
	cmd.Flags().StringVarP(&outputFile, "name", "n", "", "Name of the output file")
	cmd.Flags().StringVarP(&excludeDirs, "exclude", "x", "node_modules,dist", "Comma-separated list of directories/files/extensions to exclude")
	cmd.Flags().BoolVarP(&generateIgnore, "generate-ignore", "g", false, "Generate a default .files-combineignore file")

	return cmd
}

func parseExtensions(input string) []string {
	var exts []string
	for e := range strings.SplitSeq(input, ",") {
		e = strings.TrimSpace(strings.ToLower(e))
		if e != "" {
			if !strings.HasPrefix(e, ".") {
				e = "." + e
			}
			exts = append(exts, e)
		}
	}
	return exts
}

func hasAllowedExtension(path string, exts []string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return slices.Contains(exts, ext)
}
