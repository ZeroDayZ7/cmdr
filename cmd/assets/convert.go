package assets

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	assetconfig "github.com/zerodayz7/cmdr/internal/assets/config"
	assetconvert "github.com/zerodayz7/cmdr/internal/assets/convert"
	imageutil "github.com/zerodayz7/cmdr/internal/assets/image"
)

func NewConvertCmd() *cobra.Command {
	var inputPath string
	var outputPath string
	var outputFormat string
	var quality int
	var lossless bool
	var overwrite bool

	cmd := &cobra.Command{
		Use:   "convert",
		Short: "Convert an image between supported formats",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(inputPath) == "" {
				return fmt.Errorf("--input/-i is required")
			}
			if strings.TrimSpace(outputPath) == "" {
				return fmt.Errorf("--output/-o is required")
			}
			if quality < 1 || quality > 100 {
				return fmt.Errorf("--quality must be between 1 and 100")
			}
			if _, err := os.Stat(inputPath); err != nil {
				return fmt.Errorf("source image %s does not exist: %w", inputPath, err)
			}

			format, err := imageutil.NormalizeOutputFormat(outputPath, outputFormat)
			if err != nil {
				return err
			}
			if format != "webp" {
				if cmd.Flags().Changed("quality") {
					return fmt.Errorf("--quality is supported only for WebP output")
				}
				if lossless {
					return fmt.Errorf("--lossless is supported only for WebP output")
				}
			}

			cfg, err := assetconfig.LoadConfig("")
			if err != nil {
				return fmt.Errorf("load asset config: %w", err)
			}
			if !overwrite && cfg.Defaults.Overwrite {
				overwrite = true
			}

			resolvedFormat, err := assetconvert.ConvertImage(inputPath, outputPath, outputFormat, quality, lossless, overwrite)
			if err != nil {
				return err
			}
			cmd.Printf("Converted image saved to: %s (%s)\n", outputPath, resolvedFormat)
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputPath, "input", "i", "", "source image path")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "output image path")
	cmd.Flags().StringVar(&outputFormat, "format", "", "output format (png, jpg, jpeg, gif, or webp)")
	cmd.Flags().IntVarP(&quality, "quality", "q", 85, "WebP quality from 1 to 100")
	cmd.Flags().BoolVar(&lossless, "lossless", false, "encode WebP losslessly")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite an existing output file")
	return cmd
}
