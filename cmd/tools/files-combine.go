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
	dartOnly       bool
)

func NewFilesCombineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "files-combine",
		Aliases: []string{"fc", "combine-files"},
		Short:   "Combine contents of files with chosen extensions into one file",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if dartOnly {
				extensions = "dart"
				excludeDirs = strings.Join([]string{
					excludeDirs,
					".g.dart",
					".freezed.dart",
				}, ",")
			}

			if extensions == "" {
				extensions = "ts,tsx,jsx,json"
			}

			defaultIgnore, _ := utils.ReadIgnoreFile(".files-combineignore")
			excludeList := append(defaultIgnore, utils.ParseCommaSeparated(excludeDirs)...)

			extList := parseExtensions(extensions)

			file, err := os.Create(outputFile)
			if err != nil {
				return err
			}
			defer file.Close()

			writer := bufio.NewWriter(file)
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
					name := info.Name()

					if dartOnly && isGeneratedDartFile(name) {
						return nil
					}

					fmt.Fprintf(writer, "───────────── %s ─────────────\n", name)

					content, readErr := os.ReadFile(path)
					if readErr == nil {
						writer.Write(content)
					}
					writer.WriteString("\n\n")
				}
				return nil
			})

			if err != nil {
				return err
			}

			writer.Flush()
			fmt.Printf("Finished! Combined file: %s\n", outputFile)
			return nil
		},
	}

	cmd.Flags().StringVarP(&extensions, "extensions", "e", "", "Comma-separated list of file extensions")
	cmd.Flags().StringVarP(&outputFile, "name", "n", "", "Name of the output file")
	cmd.Flags().StringVarP(&excludeDirs, "exclude", "x", "node_modules,dist", "Excluded directories/files")
	cmd.Flags().BoolVarP(&generateIgnore, "generate-ignore", "g", false, "Generate ignore file")
	cmd.Flags().BoolVarP(&dartOnly, "dart", "d", false, "Dart mode")

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

func isGeneratedDartFile(name string) bool {
	return strings.HasSuffix(name, ".g.dart") ||
		strings.HasSuffix(name, ".freezed.dart")
}
