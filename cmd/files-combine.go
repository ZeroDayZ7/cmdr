package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	outputFile  string
	extensions  string
	excludeDirs string
)

var filesCombineCmd = &cobra.Command{
	Use:     "files-combine",
	Aliases: []string{"fc", "combine-files"},
	Short:   "Combine contents of files with chosen extensions into one file",
	Long: `Scans the current directory and all subdirectories for files with
selected extensions and combines them into a single file with headers.

Example:
  cmdr files-combine -r ts,tsx,jsx,json -n my-combined.txt
`,
	Run: func(cmd *cobra.Command, args []string) {
		if outputFile == "" {
			outputFile = "combined.txt"
		}

		if extensions == "" {
			extensions = "ts,tsx,jsx,json"
		}

		// Parse extensions into a slice
		extList := parseExtensions(extensions)
		if len(extList) == 0 {
			fmt.Println("No valid extensions provided.")
			return
		}

		fmt.Printf("Combining files with extensions: %v\n", extList)
		fmt.Printf("Output file: %s\n", outputFile)

		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Println("Error creating output file:", err)
			return
		}
		defer file.Close()

		writer := bufio.NewWriter(file)

		// Initial empty lines (4)
		for i := 0; i < 4; i++ {
			writer.WriteString("\n")
		}

		excludeList := parseExcludes(excludeDirs)

		err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() && shouldExclude(path, excludeList) {
				return filepath.SkipDir
			}

			if !info.IsDir() && hasAllowedExtension(path, extList) {
				_, fileName := filepath.Split(path)
				writer.WriteString(fmt.Sprintf("───────────── %s ─────────────\n", fileName))

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
			fmt.Println("Error while scanning files:", err)
			return
		}

		writer.Flush()
		fmt.Printf("✅ Finished! Combined file: %s\n", outputFile)
	},
}

func init() {
	filesCombineCmd.Flags().StringVarP(&extensions, "extensions", "r", "", "Comma-separated list of file extensions (e.g. ts,tsx,json)")
	filesCombineCmd.Flags().StringVarP(&outputFile, "name", "n", "", "Name of the output file")
	filesCombineCmd.Flags().StringVar(&excludeDirs, "exclude", "node_modules,dist", "Comma-separated list of directories to exclude")
	rootCmd.AddCommand(filesCombineCmd)
}

func parseExtensions(input string) []string {
	var exts []string
	for _, e := range strings.Split(input, ",") {
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

func parseExcludes(input string) []string {
	var dirs []string
	for _, d := range strings.Split(input, ",") {
		d = strings.TrimSpace(d)
		if d != "" {
			dirs = append(dirs, d)
		}
	}
	return dirs
}

func shouldExclude(path string, excludeList []string) bool {
	for _, ex := range excludeList {
		if strings.Contains(path, string(filepath.Separator)+ex) {
			return true
		}
	}
	return false
}

func hasAllowedExtension(path string, exts []string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, e := range exts {
		if e == ext {
			return true
		}
	}
	return false
}
