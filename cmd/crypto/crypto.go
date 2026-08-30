package crypto

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "crypto",
		Aliases: []string{"c"},
		Short:   "Operacje kryptograficzne: klucze, szyfrowanie, hashowanie",
	}

	cmd.AddCommand(
		newGenerateKeyCmd(),
		newEncryptCmd(),
		newDecryptCmd(),
		newPassEncryptCmd(),
		newPassDecryptCmd(),
		// newHashCmd(),
		newRandomCmd(),
	)

	return cmd
}
