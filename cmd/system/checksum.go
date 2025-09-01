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
)

func NewChecksumCmd() *cobra.Command {
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
  cmdr checksum example.txt -o -a sha512 # auto-generate file example.sha512
  cmdr checksum example.txt -f mysum.txt # custom output file
  cmdr checksum example.txt -a md5
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			// otwieranie pliku do hashowania
			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("error opening file %s: %w", filePath, err)
			}
			defer file.Close()

			// wybór algorytmu
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

			if _, err := io.Copy(hasher, file); err != nil {
				return fmt.Errorf("error reading file %s: %w", filePath, err)
			}

			checksum := fmt.Sprintf("%x", hasher.Sum(nil))

			// ustalanie pliku wyjściowego
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
				fmt.Printf("✅ Checksum saved to %s\n", outPath)
			} else {
				fmt.Printf("%s  %s\n", checksum, filePath)
			}

			return nil
		},
	}

	// flagi
	cmd.Flags().BoolVarP(&autoOutput, "output", "o", false, "Generate checksum file automatically")
	cmd.Flags().StringVarP(&outputFile, "file", "f", "", "Custom output filename")
	cmd.Flags().StringVarP(&algo, "algo", "a", "sha256", "Algorithm: sha256|sha512|md5")

	return cmd
}
