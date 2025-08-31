package crypto

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/crypto"
)

var (
	decryptFile string
	decryptDir  string
	decryptKey  string
	decryptOut  string
)

func newDecryptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "decrypt",
		Aliases: []string{"dec"},
		Short:   "Decrypt a file or folder (AES)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if decryptFile == "" && decryptDir == "" {
				return fmt.Errorf("please provide --file or --dir")
			}
			if decryptOut == "" {
				decryptOut = "decrypted"
			}
			if err := os.MkdirAll(decryptOut, 0700); err != nil {
				return fmt.Errorf("failed to create output folder: %w", err)
			}
			if decryptKey == "" {
				return fmt.Errorf("please provide AES key with --key")
			}

			key, err := hex.DecodeString(decryptKey)
			if err != nil {
				return fmt.Errorf("failed to decode AES key: %w", err)
			}

			if decryptFile != "" {
				outPath := decryptOut + "/" + filepath.Base(decryptFile)
				// usuń końcówkę .enc
				if len(outPath) > 4 && outPath[len(outPath)-4:] == ".enc" {
					outPath = outPath[:len(outPath)-4]
				}
				if err := crypto.DecryptFileAES(decryptFile, key, outPath); err != nil {
					return err
				}

				fmt.Println("Decrypted file:", outPath)
			} else if decryptDir != "" {
				if err := crypto.DecryptFolderAES(decryptDir, key, decryptOut); err != nil {
					return err
				}
				fmt.Println("Folder decrypted successfully to:", decryptOut)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&decryptFile, "file", "f", "", "File to decrypt")
	cmd.Flags().StringVarP(&decryptDir, "dir", "d", "", "Folder to decrypt (.zip.enc)")
	cmd.Flags().StringVarP(&decryptKey, "key", "k", "", "AES key in hex")
	cmd.Flags().StringVarP(&decryptOut, "out", "o", "", "Output folder (default: decrypted)")

	return cmd
}
