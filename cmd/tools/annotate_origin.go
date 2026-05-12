package tools

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/annotate"
)

func NewAnnotateCmd() *cobra.Command {
	var filePath string
	var dirPath string
	var dryRun bool
	var verbose bool

	cmd := &cobra.Command{
		Use:     "annotate",
		Aliases: []string{"ann", "ant", "at"},
		Short:   "Insert project path annotations into files",
		RunE: func(cmd *cobra.Command, args []string) error {

			if filePath == "" && dirPath == "" {
				dirPath = "."
			}

			cfg := annotate.Config{
				DryRun:  dryRun,
				Verbose: verbose,
			}

			if filePath != "" {
				return annotate.AnnotateFile(filePath, cfg)
			}

			return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() {
					return nil
				}

				return annotate.AnnotateFile(path, cfg)
			})
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "File to annotate")
	cmd.Flags().StringVarP(&dirPath, "dir", "d", "", "Directory to annotate")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	return cmd
}
