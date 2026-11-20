package assets

import (
	"github.com/spf13/cobra"
)

func NewAssetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assets",
		Short: "Generate and process image assets",
		Long:  "Create favicon sets and resize image assets for web and application projects.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewFaviconCmd())
	cmd.AddCommand(NewResizeCmd())
	cmd.AddCommand(NewConvertCmd())
	cmd.AddCommand(NewConfigCmd())
	return cmd
}
