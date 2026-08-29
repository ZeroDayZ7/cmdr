package bootstrap

import "github.com/spf13/cobra"

func NewBootstrapCmd() *cobra.Command {
	bootstrapCmd := &cobra.Command{
		Use:     "bootstrap",
		Aliases: []string{"bs"},
		Short:   "Narzędzia do przygotowywania i bootstrapowania konfiguracji",
	}

	// Rejestracja podkomendy prepare-password
	bootstrapCmd.AddCommand(NewPreparePasswordCmd())

	return bootstrapCmd
}
