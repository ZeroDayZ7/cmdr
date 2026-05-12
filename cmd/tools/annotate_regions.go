package tools

import (
	"github.com/spf13/cobra"
	annotate "github.com/zerodayz7/cmdr/internal/annotate_region"
	"github.com/zerodayz7/cmdr/internal/logger" // Import musi być użyty
)

var (
	regDir     string
	regFile    string
	regDryRun  bool
	regVerbose bool
	// Inicjalizujemy logger, aby był dostępny w całym pliku
	log = &logger.ConsoleLogger{}
)

func NewAnnotateRegionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "annotate-regions",
		Aliases: []string{"reg", "ann"},
		Short:   "Wraps functions with #region tags",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := annotate.Config{
				DryRun:  regDryRun,
				Verbose: regVerbose,
			}

			target := regDir
			if regFile != "" {
				target = regFile
			}
			if target == "" {
				target = "."
			}

			if cfg.DryRun {
				// Teraz 'log' jest zdefiniowany powyżej
				log.Info("Running in DRY-RUN mode. No files will be modified.")
			}

			err := annotate.Process(target, cfg)
			if err != nil {
				// Log błędu przed wyjściem
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
