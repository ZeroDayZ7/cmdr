package tools

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/annotate"
	region "github.com/zerodayz7/cmdr/internal/annotate_region"
	"github.com/zerodayz7/cmdr/internal/logger"
	"github.com/zerodayz7/cmdr/internal/profiles"
)

var (
	regDir      string
	regFile     string
	regProfile  string
	regDryRun   bool
	regVerbose bool
	log         = &logger.ConsoleLogger{}
)

func NewAnnotateRegionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "annotate-regions",
		Aliases: []string{"reg", "ann"},
		Short:   "Wraps functions with #region tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Ładowanie konfiguracji profilów
			profilesConfig, err := profiles.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load profiles: %w", err)
			}

			target := regDir
			if regFile != "" {
				target = regFile
			}
			if target == "" {
				target = "."
			}

			// 2. Wykrywanie lub dopasowywanie profilu
			absTarget, _ := filepath.Abs(target)
			var activeProfile *profiles.Profile

			if regProfile != "" {
				for _, p := range profilesConfig.Profiles {
					if p.Name == regProfile {
						pCopy := p
						activeProfile = &pCopy
						break
					}
				}
			} else {
				activeProfile = profiles.DetectProjectProfile(absTarget, profilesConfig)
			}

			consoleLogger := &logger.ConsoleLogger{IsVerbose: regVerbose}

			// 3. Poprawne przekazanie ProfilesConfig do obiektu Config
			cfg := annotate.Config{
				DryRun:         regDryRun,
				Verbose:        regVerbose,
				Profile:        activeProfile,
				ProfilesConfig: profilesConfig,
				Log:            consoleLogger,
			}

			if cfg.DryRun {
				consoleLogger.Info("Running in DRY-RUN mode. No files will be modified.")
			}

			err = region.Process(target, cfg, cmd.Context())
			if err != nil {
				consoleLogger.Error("Annotation failed: %v", err)
				return err
			}

			consoleLogger.Success("Process finished successfully.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&regDir, "directory", "d", "", "Directory to process")
	cmd.Flags().StringVarP(&regFile, "file", "f", "", "Specific file to process")
	cmd.Flags().StringVarP(&regProfile, "profile", "p", "", "Use specific profile (e.g. go, typescript, flutter)")
	cmd.Flags().BoolVar(&regDryRun, "dry-run", false, "Preview changes without modifying files")
	cmd.Flags().BoolVarP(&regVerbose, "verbose", "v", false, "Print detailed information")

	return cmd
}