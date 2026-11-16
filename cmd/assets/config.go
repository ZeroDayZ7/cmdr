package assets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	assetconfig "github.com/zerodayz7/cmdr/internal/assets/config"
)

func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage the asset configuration",
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
		RunE: func(cmd *cobra.Command, args []string) error {
			path := customPath
			if path == "" {
				resolved, err := assetconfig.GetConfigPath()
				if err != nil {
					return fmt.Errorf("resolve default config path: %w", err)
				}
				path = resolved
			}

			if !force {
				if _, err := os.Stat(path); err == nil {
					return fmt.Errorf("configuration already exists at %s; use --force to overwrite", path)
				}
			}

			if err := assetconfig.InitConfig(path, force); err != nil {
				return err
			}
			fmt.Printf("Initialized asset configuration at: %s\n", path)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite an existing configuration file")
	cmd.Flags().StringVar(&customPath, "path", "", "custom config path (defaults to ~/.cmdr/assets.json)")
	cmd.Flags().StringVarP(&customPath, "output", "o", "", "custom config path (defaults to ~/.cmdr/assets.json)")
	cmd.MarkFlagsMutuallyExclusive("path", "output")
	return cmd
}

func normalizeAssetPath(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	if filepath.IsAbs(raw) {
		return raw
	}
	return filepath.Clean(raw)
}
