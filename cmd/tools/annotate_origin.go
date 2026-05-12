package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	annotate "github.com/zerodayz7/cmdr/internal/annotate_origin"
	"github.com/zerodayz7/cmdr/internal/profiles"
)

func NewAnnotateCmd() *cobra.Command {
	var filePath string
	var dirPath string
	var profileName string
	var dryRun bool
	var verbose bool

	cmd := &cobra.Command{
		Use:     "annotate",
		Aliases: []string{"ann", "ant", "at"},
		Short:   "Insert project path annotations into files",
		RunE: func(cmd *cobra.Command, args []string) error {
			profilesConfig, err := profiles.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load profiles: %w", err)
			}

			var activeProfile *profiles.Profile
			if profileName != "" {
				for _, p := range profilesConfig.Profiles {
					if p.Name == profileName {
						profileCopy := p
						activeProfile = &profileCopy
						break
					}
				}
				if activeProfile == nil {
					return fmt.Errorf("profile %q not found", profileName)
				}
			}

			cfg := annotate.Config{
				DryRun:         dryRun,
				Verbose:        verbose,
				Profile:        activeProfile,
				ProfilesConfig: profilesConfig,
			}

			if filePath == "" && dirPath == "" {
				dirPath = "."
			}

			if filePath != "" {
				return annotate.AnnotateFile(filePath, cfg)
			}

			return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				return annotate.AnnotateFile(path, cfg)
			})
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "File to annotate")
	cmd.Flags().StringVarP(&dirPath, "dir", "d", "", "Directory to annotate")
	cmd.Flags().StringVarP(&profileName, "profile", "p", "", "Use specific profile (e.g. go, rust, flutter)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	return cmd
}
