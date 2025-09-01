package info

import (
	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/cmd/info/info_go"
)

func NewInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "info",
		Aliases: []string{"i"},
		Short:   "Show recommended project structures and use cases",
	}

	cmd.AddCommand(
		info_go.NewGoCmd(),
	)

	return cmd
}
