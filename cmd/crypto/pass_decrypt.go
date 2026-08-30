package crypto

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/crypto"
)

var (
	passDecFile string
	passDecPass string
	passDecOut  string
)

func newPassDecryptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pass-decrypt",
		Aliases: []string{"pd"},
		Short:   "Decrypt a file encrypted with Argon2id",
		RunE: func(cmd *cobra.Command, args []string) error {
			if passDecFile == "" {
				return fmt.Errorf("please provide a file to decrypt (--file / -f)")
			}
			if passDecPass == "" {
				return fmt.Errorf("please provide a password (--pass / -p)")
			}

			if passDecOut == "" {
				passDecOut = "decrypted"
			}

			if err := os.MkdirAll(passDecOut, 0700); err != nil {
				return fmt.Errorf("failed to create output folder: %w", err)
			}

			outPath := filepath.Join(passDecOut, filepath.Base(passDecFile))
			if len(outPath) > 4 && outPath[len(outPath)-4:] == ".enc" {
				outPath = outPath[:len(outPath)-4]
			}

			if err := crypto.DecryptFileArgon2(passDecFile, passDecPass, outPath); err != nil {
				return fmt.Errorf("decryption failed: %w", err)
			}

			fmt.Println("File decrypted successfully:")
			fmt.Println("  Output file:", outPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&passDecFile, "file", "f", "", "File to decrypt (required)")
	cmd.Flags().StringVarP(&passDecPass, "pass", "p", "", "Password for Argon2id decryption (required)")
	cmd.Flags().StringVarP(&passDecOut, "out", "o", "", "Output folder (default: decrypted)")

	return cmd
}
