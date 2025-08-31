package crypto

import (
	"github.com/spf13/cobra"
)

// NewCmd tworzy główną komendę `crypto`
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crypto",
		Short: "Operacje kryptograficzne: klucze, szyfrowanie, hashowanie",
	}

	// Dodaj podkomendy
	cmd.AddCommand(
		newGenerateKeyCmd(),
		newEncryptCmd(),
		newDecryptCmd(),
		// newHashCmd(),
		newRandomCmd(),
	)

	return cmd
}
