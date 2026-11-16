package assets

import (
	"fmt"

	"github.com/spf13/cobra"
	assetconfig "github.com/zerodayz7/cmdr/internal/assets/config"
)

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage the asset configuration",
		Args:  cobra.NoArgs,
	}
	cmd.AddCommand(NewConfigInitCmd())
	return cmd
}

func NewConfigInitCmd() *cobra.Command {
	var force bool
	var customPath string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the default assets.json configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := customPath
			if path == "" {
				resolved, err := assetconfig.GetConfigPath()
				if err != nil {
					return fmt.Errorf("resolve default config path: %w", err)
				}
				path = resolved
			}

			if err := assetconfig.InitConfig(path, force); err != nil {
				return err
			}
			cmd.Printf("Initialized asset configuration at: %s\n", path)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite an existing configuration file")
	cmd.Flags().StringVarP(&customPath, "path", "p", "", "custom config path (defaults to ~/.cmdr/assets.json)")
	return cmd
}
