package system

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/logger"
)

func NewChecksumCmd(l logger.Logger) *cobra.Command {
	var autoOutput bool
	var outputFile string
	var algo string

	cmd := &cobra.Command{
		Use:     "checksum [file]",
		Aliases: []string{"csum", "sha"},
		Short:   "Compute checksum of a file",
		Long: `This command calculates the checksum of a given file.
Examples:
  cmdr checksum example.txt          # prints checksum to console
  cmdr checksum example.txt -o       # auto-generate file example.sha256
  cmdr checksum example.txt -f out.txt # custom output file
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			ctx := cmd.Context()

			l.Debug("Starting checksum calculation", "file", filePath, "algo", algo)

			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("error opening file %s: %w", filePath, err)
			}
			defer file.Close()

			var hasher hash.Hash
			switch strings.ToLower(algo) {
			case "sha256", "":
				hasher = sha256.New()
			case "sha512":
				hasher = sha512.New()
			case "md5":
				hasher = md5.New()
			default:
				return fmt.Errorf("unsupported algorithm: %s", algo)
			}

			errChan := make(chan error, 1)
			go func() {
				_, err := io.Copy(hasher, file)
				errChan <- err
			}()

			select {
			case <-ctx.Done():
				l.Error("Operation cancelled by user")
				return ctx.Err()
			case err := <-errChan:
				if err != nil {
					return fmt.Errorf("error reading file %s: %w", filePath, err)
				}
			}

			checksum := fmt.Sprintf("%x", hasher.Sum(nil))

			outPath := ""
			if outputFile != "" {
				outPath = outputFile
			} else if autoOutput {
				ext := strings.ToLower(algo)
				if ext == "" {
					ext = "sha256"
				}
				outPath = filePath + "." + ext
			}

			if outPath != "" {
				if !filepath.IsAbs(outPath) {
					outPath = filepath.Join(".", outPath)
				}
				if err := os.WriteFile(outPath, []byte(checksum+"\n"), 0644); err != nil {
					return fmt.Errorf("error writing to file %s: %w", outPath, err)
				}
				l.Success("Checksum saved to %s", outPath)
			} else {
				l.Success("%s  %s", checksum, filePath)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&autoOutput, "output", "o", false, "Generate checksum file automatically")
	cmd.Flags().StringVarP(&outputFile, "file", "f", "", "Custom output filename")
	cmd.Flags().StringVarP(&algo, "algo", "a", "sha256", "Algorithm: sha256|sha512|md5")

	return cmd
}
