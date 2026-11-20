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
			cmd.Printf("%s v%s\n", config.Name, config.Version)
			if config.GitCommit != "" && config.GitCommit != "unknown" {
				commit := config.GitCommit
				if len(commit) > 12 {
					commit = commit[:12]
				}
				cmd.Printf("commit: %s\n", commit)
			}
			if config.BuildDate != "" && config.BuildDate != "unknown" {
				cmd.Printf("built: %s\n", config.BuildDate)
			}
		},
	}
}
