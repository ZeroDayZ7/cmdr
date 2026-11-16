package assets

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	assetconfig "github.com/zerodayz7/cmdr/internal/assets/config"
	"github.com/zerodayz7/cmdr/internal/assets/favicon"
)

func NewFaviconCmd() *cobra.Command {
	var inputPath string
	var outputDir string
	var sizesFlag string
	var overwrite bool

	cmd := &cobra.Command{
		Use:   "favicon",
		Short: "Generate a favicon set from a single source image",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(inputPath) == "" {
				return fmt.Errorf("--input/-i is required")
			}
			if _, err := os.Stat(inputPath); err != nil {
				return fmt.Errorf("input image %s does not exist: %w", inputPath, err)
			}

			cfg, err := assetconfig.LoadConfig("")
			if err != nil {
				return fmt.Errorf("load asset config: %w", err)
			}
			if strings.TrimSpace(outputDir) == "" {
				outputDir = cfg.Defaults.OutputDir
			}
			if strings.TrimSpace(outputDir) == "" {
				outputDir = "generated"
			}

			finalSizes, err := parseSizes(sizesFlag)
			if err != nil {
				return err
			}
			if len(finalSizes) == 0 {
				if preset, ok := cfg.Presets["favicon"]; ok && len(preset.Sizes) > 0 {
					finalSizes = preset.Sizes
				} else {
					finalSizes = favicon.DefaultSizes()
				}
			}
			if !overwrite && cfg.Defaults.Overwrite {
				overwrite = true
			}
			preserveTransparency := true
			if preset, ok := cfg.Presets["favicon"]; ok {
				preserveTransparency = preset.PreserveTransparency
			}

			files, err := favicon.GenerateFaviconSetWithOptions(inputPath, outputDir, finalSizes, overwrite, preserveTransparency)
			if err != nil {
				return err
			}
			for _, name := range files {
				cmd.Printf("Created: %s\n", name)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputPath, "input", "i", "", "source image path")
	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "output directory for generated files")
	cmd.Flags().StringVar(&sizesFlag, "sizes", "", "comma-separated target favicon sizes, e.g. 16,32,48")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite existing generated files")
	return cmd
}

func parseSizes(raw string) ([]int, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	result := make([]int, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			return nil, fmt.Errorf("invalid favicon size value: empty item")
		}
		value, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, fmt.Errorf("invalid favicon size value %q: %w", trimmed, err)
		}
		if value <= 0 {
			return nil, fmt.Errorf("invalid favicon size value %q: expected a positive integer", trimmed)
		}
		if value > 2048 {
			return nil, fmt.Errorf("invalid favicon size value %q: exceeds supported maximum 2048", trimmed)
		}
		result = append(result, value)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("favicon sizes list is empty")
	}
	return result, nil
}
