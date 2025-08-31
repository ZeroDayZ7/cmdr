package crypto

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/crypto"
)

var (
	encryptFile string
	encryptDir  string
	encryptKey  string
	outFolder   string
)

func newEncryptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "encrypt",
		Aliases: []string{"enc"},
		Short:   "Encrypt a file or folder (AES)",
		Long: `
Encrypt a single file or a folder. 
Encrypted file and AES key are saved in the output folder.
If no key is provided, a random AES key is generated.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if encryptFile == "" && encryptDir == "" {
				return fmt.Errorf("please provide --file or --dir")
			}
			if outFolder == "" {
				outFolder = "encrypted"
			}
			if err := os.MkdirAll(outFolder, 0700); err != nil {
				return fmt.Errorf("failed to create output folder: %w", err)
			}
			var key []byte
			var err error
			if encryptKey == "" {
				key, err = crypto.GenerateRandomAESKey(32)
				if err != nil {
					return err
				}
				encryptKey = hex.EncodeToString(key)
				fmt.Println("Generated AES key:", encryptKey)
			} else {
				key, err = hex.DecodeString(encryptKey)
				if err != nil {
					return fmt.Errorf("failed to decode AES key: %w", err)
				}
			}
			if encryptFile != "" {
				encPath, keyPath, err := crypto.EncryptFileAES(encryptFile, key, outFolder)
				if err != nil {
					return err
				}
				fmt.Println("Encrypted file:", encPath)
				fmt.Println("AES key saved:", keyPath)
			} else if encryptDir != "" {
				encPath, keyPath, err := crypto.EncryptFolderAES(encryptDir, key, outFolder)
				if err != nil {
					return err
				}
				fmt.Println("Encrypted folder:", encPath)
				fmt.Println("AES key saved:", keyPath)

			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&encryptFile, "file", "f", "", "File to encrypt")
	cmd.Flags().StringVarP(&encryptDir, "dir", "d", "", "Folder to encrypt")
	cmd.Flags().StringVarP(&encryptKey, "key", "k", "", "AES key in hex (optional)")
	cmd.Flags().StringVarP(&outFolder, "out", "o", "", "Output folder (default: encrypted)")
	return cmd
}
