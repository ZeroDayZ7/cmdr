package cmd

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var hashdirCmd = &cobra.Command{
	Use:     "hashdir [dir]",
	Aliases: []string{"hd", "dirhash"},
	Short:   "Compute SHA256 hash of a directory",
	Long: `This command recursively calculates the SHA256 hash of all files in a directory.
You can exclude certain subdirectories using --exclude, and optionally write output to a file.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir := args[0]

		excludeDirs, _ := cmd.Flags().GetStringArray("exclude")
		outputFile, _ := cmd.Flags().GetString("out")

		files := []string{}

		// Walk directory recursively
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				for _, ex := range excludeDirs {
					if strings.HasPrefix(path, filepath.Join(dir, ex)) {
						return filepath.SkipDir
					}
				}
				return nil
			}

			files = append(files, path)
			return nil
		})
		if err != nil {
			fmt.Printf("Error walking directory: %v\n", err)
			return
		}

		sort.Strings(files) // aby kolejność plików nie zmieniała hash

		hash := sha256.New()
		for _, file := range files {
			f, err := os.Open(file)
			if err != nil {
				fmt.Printf("Error opening file %s: %v\n", file, err)
				return
			}
			if _, err := io.Copy(hash, f); err != nil {
				fmt.Printf("Error reading file %s: %v\n", file, err)
				f.Close()
				return
			}
			f.Close()
		}

		sum := fmt.Sprintf("%x", hash.Sum(nil))
		fmt.Printf("Directory hash: %s\n", sum)

		if outputFile != "" {
			if err := os.WriteFile(outputFile, []byte(sum), 0644); err != nil {
				fmt.Printf("Error writing to file %s: %v\n", outputFile, err)
			} else {
				fmt.Printf("Hash written to %s\n", outputFile)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(hashdirCmd)
	hashdirCmd.Flags().StringArrayP("exclude", "e", []string{}, "Directories to exclude from hashing")
	hashdirCmd.Flags().StringP("out", "o", "", "Output file to save the hash")
}
