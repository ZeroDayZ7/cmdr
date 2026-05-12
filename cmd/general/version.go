package general

import (
	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/config"
	"github.com/zerodayz7/cmdr/internal/logger"
)

func NewVersionCmd(l logger.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Shows the CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			l.Success("%s v%s", config.Name, config.Version)

			// -v
			l.Debug("Contact: %s", config.Contact)
		},
	}
}
