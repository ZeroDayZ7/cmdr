package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// #region NewCleanCmd
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
				fmt.Println("Please provide a directory name to delete using the -d flag")
				return
			}

			fmt.Printf("Cleaning all '%s' folders in the project...\n", dirName)

			err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() && info.Name() == dirName {
					fmt.Printf("Deleting: %s\n", path)
					if rmErr := os.RemoveAll(path); rmErr != nil {
						fmt.Printf("Error deleting %s: %v\n", path, rmErr)
					}
				}
				return nil
			})

			if err != nil {
				fmt.Println("Error walking the path:", err)
				return
			}

			fmt.Println("Done!")
		},
	}

	cmd.Flags().StringVarP(&dirName, "dir", "d", "", "Name of the directory to delete")
	return cmd
}
