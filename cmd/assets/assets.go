package assets

import (
	"github.com/spf13/cobra"
)

func NewAssetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assets",
		Short: "Generate and process image assets",
		Long:  "Create favicon sets, resize or convert images, and create PDFs from numbered image files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewFaviconCmd())
	cmd.AddCommand(NewResizeCmd())
	cmd.AddCommand(NewConvertCmd())
	cmd.AddCommand(NewPDFCmd())
	cmd.AddCommand(NewConfigCmd())
	return cmd
}
