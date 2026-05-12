package tools

import (
	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/annotate" // Import współdzielonego pakietu
	region "github.com/zerodayz7/cmdr/internal/annotate_region"
	"github.com/zerodayz7/cmdr/internal/logger"
)

var (
	regDir     string
	regFile    string
	regDryRun  bool
	regVerbose bool
	log        = &logger.ConsoleLogger{}
)

func NewAnnotateRegionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "annotate-regions",
		Aliases: []string{"reg", "ann"},
		Short:   "Wraps functions with #region tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Używamy typu Config z pakietu shared (annotate)
			cfg := annotate.Config{
				DryRun:  regDryRun,
				Verbose: regVerbose,
				Log:     log, // Przekazujemy logger do konfiguracji
			}

			target := regDir
			if regFile != "" {
				target = regFile
			}
			if target == "" {
				target = "."
			}

			if cfg.DryRun {
				log.Info("Running in DRY-RUN mode. No files will be modified.")
			}

			err := region.Process(target, cfg, cmd.Context())
			if err != nil {
				log.Error("Annotation failed: %v", err)
				return err
			}

			log.Success("Process finished successfully.")
			return nil
		},
	}

	cmd.Flags().StringVarP(&regDir, "directory", "d", "", "Directory to process")
	cmd.Flags().StringVarP(&regFile, "file", "f", "", "Specific file to process")
	cmd.Flags().BoolVar(&regDryRun, "dry-run", false, "Preview changes without modifying files")
	cmd.Flags().BoolVarP(&regVerbose, "verbose", "v", false, "Print detailed information")

	return cmd
}
