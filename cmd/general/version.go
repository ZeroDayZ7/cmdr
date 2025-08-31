package general

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Shows the CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("cmdr v2.1.4")
		},
	}
}
