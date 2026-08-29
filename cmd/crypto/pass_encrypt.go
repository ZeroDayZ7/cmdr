// cmd/crypto/pass_encrypt.go
package crypto

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/crypto"
)

var (
	passEncFile string
	passEncPass string
	passEncOut  string
)

func newPassEncryptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pass-encrypt",
		Aliases: []string{"penc"},
		Short:   "Encrypt a file using a password (Argon2id + AES-GCM)",
		Long: `
Encrypts a single file using a user-provided password.
Derives a 256-bit key using Argon2id and seals the file with AES-GCM.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if passEncFile == "" {
				return fmt.Errorf("please provide a file to encrypt (--file / -f)")
			}
			if passEncPass == "" {
				return fmt.Errorf("please provide a password (--pass / -p)")
			}

			if passEncOut == "" {
				passEncOut = "encrypted"
			}

			if err := os.MkdirAll(passEncOut, 0700); err != nil {
				return fmt.Errorf("failed to create output folder: %w", err)
			}

			encPath, err := crypto.EncryptFileArgon2(passEncFile, passEncPass, passEncOut)
			if err != nil {
				return fmt.Errorf("encryption failed: %w", err)
			}

			fmt.Println("File encrypted successfully using Argon2id:")
			fmt.Println("  Output file:", encPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&passEncFile, "file", "f", "", "File to encrypt (required)")
	cmd.Flags().StringVarP(&passEncPass, "pass", "p", "", "Password for Argon2id key derivation (required)")
	cmd.Flags().StringVarP(&passEncOut, "out", "o", "", "Output folder (default: encrypted)")

	return cmd
}
