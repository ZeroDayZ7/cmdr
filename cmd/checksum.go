package cmd

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var checksumCmd = &cobra.Command{
	Use:     "checksum [file]",
	Aliases: []string{"csum", "sha"},
	Short:   "Compute SHA256 checksum of a file",
	Long:    `This command calculates the SHA256 checksum of a given file.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Error opening file %s: %v\n", filePath, err)
			return
		}
		defer file.Close()

		hash := sha256.New()
		if _, err := io.Copy(hash, file); err != nil {
			fmt.Printf("Error reading file %s: %v\n", filePath, err)
			return
		}

		checksum := fmt.Sprintf("%x", hash.Sum(nil))
		fmt.Printf("SHA256 checksum of %s: %s\n", filePath, checksum)
	},
}

func init() {
	rootCmd.AddCommand(checksumCmd)
}
