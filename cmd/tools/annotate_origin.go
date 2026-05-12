package tools

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	annotate "github.com/zerodayz7/cmdr/internal/annotate_origin"
	"github.com/zerodayz7/cmdr/internal/profiles"
	"github.com/zerodayz7/cmdr/internal/ui"
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

			logger := &ui.ConsoleLogger{IsVerbose: verbose}

			cfg := annotate.Config{
				DryRun:         dryRun,
				Verbose:        verbose,
				Profile:        activeProfile,
				ProfilesConfig: profilesConfig,
				Log:            logger,
			}

			var paths []string
			if filePath != "" {
				paths = append(paths, filePath)
			} else {
				if dirPath == "" {
					dirPath = "."
				}
				err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !info.IsDir() {
						paths = append(paths, path)
					}
					return nil
				})
				if err != nil {
					return fmt.Errorf("error walking path: %w", err)
				}
			}

			if verbose {
				logger.Info("Starting annotation process for %d files...", len(paths))
			}

			results := annotate.ProcessBatch(cmd.Context(), paths, cfg)

			var errCount int
			for _, res := range results {
				if res.Err != nil {
					errCount++
					logger.Error("%s: %v", res.Path, res.Err)
				}
			}

			if verbose {
				fmt.Printf("\nFinished: %d tasks, %d errors\n", len(results), errCount)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&filePath, "file", "f", "", "File to annotate")
	cmd.Flags().StringVarP(&dirPath, "dir", "d", "", "Directory to annotate")
	cmd.Flags().StringVarP(&profileName, "profile", "p", "", "Use specific profile (e.g. go, rust, flutter)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without writing")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	return cmd
}
