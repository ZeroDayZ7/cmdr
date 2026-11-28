package crypto

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/crypto"
)

var (
	keyType string
	keySize int
	outDir  string
)

// newGenerateKeyCmd returns the generate-key subcommand
func newGenerateKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate-key",
		Aliases: []string{"gen-key", "gkey", "gk"},
		Short:   "Generate an RSA key pair",
		Long: `
Generate an RSA private/public key pair and save it in a folder.
Defaults to creating a 'keys' folder in the current directory.
Supported types: rsa (2048, 3072, or 4096 bits).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Set default folder if not provided
			if outDir == "" {
				outDir = "keys"
			}

			// Create folder if it doesn't exist
			if err := os.MkdirAll(outDir, 0700); err != nil {
				return fmt.Errorf("failed to create folder: %w", err)
			}

			privPath := filepath.Join(outDir, "private.pem")
			pubPath := filepath.Join(outDir, "public.pem")

			switch keyType {
			case "rsa":
				if keySize != 2048 && keySize != 3072 && keySize != 4096 {
					return fmt.Errorf("invalid RSA key size: %d (use 2048, 3072, 4096)", keySize)
				}
				err := crypto.GenerateRSAKey(keySize, privPath, pubPath)
				if err == nil {
					fmt.Printf("RSA key pair generated successfully in folder: %s\n", outDir)
				}
				return err
			default:
				return fmt.Errorf("unsupported key type: %s", keyType)
			}
		},
	}

	// Flags
	cmd.Flags().StringVarP(&keyType, "type", "t", "rsa", "Key type (currently supported: rsa)")
	cmd.Flags().IntVarP(&keySize, "size", "s", 4096, "Key size in bits (RSA only: 2048, 3072, 4096)")
	cmd.Flags().StringVarP(&outDir, "dir", "d", "", "Output folder for key pair (default: keys)")

	return cmd
}
