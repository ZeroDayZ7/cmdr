package tools

import (
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewCleanCmd() *cobra.Command {
	var dirName string

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Deletes all directories with the given name in the current project",
		Long: `Example usage:
  cmdr clean -d logs
Deletes all directories named "logs" in the current directory and subdirectories.`,
		Run: func(cmd *cobra.Command, args []string) {
			if dirName == "" {
				slog.Warn("No directory name provided. Use -d flag.")
				return
			}

			slog.Info("Cleaning folders", "target", dirName)

			err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					slog.Error("Error accessing path", "path", path, "error", err)
					return err
				}

				if d.IsDir() && d.Name() == dirName {
					slog.Info("Deleting directory", "path", path)

					if rmErr := os.RemoveAll(path); rmErr != nil {
						slog.Error("Failed to delete directory", "path", path, "error", rmErr)
					}

					return filepath.SkipDir
				}

				return nil
			})

			if err != nil {
				slog.Error("Cleanup process failed", "error", err)
				os.Exit(1)
			}

			slog.Info("Cleanup completed successfully")
		},
	}

	cmd.Flags().StringVarP(&dirName, "dir", "d", "", "Name of the directory to delete")
	return cmd
}
