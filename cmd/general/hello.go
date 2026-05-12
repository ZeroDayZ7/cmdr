package general

import (
	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/logger"
)

func NewHelloCmd(l logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "hello",
		Short: "Say hello",
		Long:  `This command is used to test write out a greeting in the CLI.`,
		Run: func(cmd *cobra.Command, args []string) {
			l.Success("Hello from cmdr! System is operational.")

			// -v
			l.Debug("Hello command executed successfully")
		},
	}
}
