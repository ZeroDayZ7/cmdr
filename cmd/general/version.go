package general

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/config"
	"github.com/zerodayz7/cmdr/internal/logger"
)

func NewVersionCmd(l logger.Logger) *cobra.Command {
	var shortOutput bool
	var jsonOutput bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Shows the CLI version and build information",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			info := config.GetBuildInfo()
			if shortOutput {
				cmd.Println(info.Version)
				return nil
			}
			if jsonOutput {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetIndent("", "  ")
				return encoder.Encode(info)
			}

			cmd.Printf("%s v%s\n\n", info.Name, info.Version)
			cmd.Printf("Version:    %s\n", info.Version)
			cmd.Printf("Git commit: %s\n", info.GitCommit)
			cmd.Printf("Build date: %s\n", info.BuildDate)
			cmd.Printf("Platform:   %s\n", info.Platform)
			cmd.Printf("Go version: %s\n", info.GoVersion)
			return nil
		},
	}
	cmd.Flags().BoolVar(&shortOutput, "short", false, "print only the version")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "print build information as JSON")
	return cmd
}
