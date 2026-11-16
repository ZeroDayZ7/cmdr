package assets

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	assetconfig "github.com/zerodayz7/cmdr/internal/assets/config"
	assetresize "github.com/zerodayz7/cmdr/internal/assets/resize"
)

func NewResizeCmd() *cobra.Command {
	var inputPath string
	var outputPath string
	var width int
	var height int
	var fit string
	var overwrite bool

	cmd := &cobra.Command{
		Use:   "resize",
		Short: "Resize a source image while keeping aspect ratio or stretching it",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(inputPath) == "" {
				return fmt.Errorf("--input/-i is required")
			}
			if strings.TrimSpace(outputPath) == "" {
				return fmt.Errorf("--output/-o is required")
			}
			if _, err := os.Stat(inputPath); err != nil {
				return fmt.Errorf("source image %s does not exist: %w", inputPath, err)
			}
			cfg, err := assetconfig.LoadConfig("")
			if err != nil {
				return fmt.Errorf("load asset config: %w", err)
			}
			if !overwrite && cfg.Defaults.Overwrite {
				overwrite = true
			}
			if width <= 0 || height <= 0 {
				return fmt.Errorf("--width/-w and --height/-H must be positive integers")
			}
			if fit == "" {
				fit = "contain"
			}
			switch strings.ToLower(fit) {
			case "contain", "cover", "stretch":
			default:
				return fmt.Errorf("invalid fit mode %q: expected contain, cover, or stretch", fit)
			}
			if err := assetresize.ResizeImage(inputPath, outputPath, width, height, fit, overwrite); err != nil {
				return err
			}
			cmd.Printf("Resized image saved to: %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&inputPath, "input", "i", "", "source image path")
	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "output image path")
	cmd.Flags().IntVarP(&width, "width", "w", 0, "target width")
	cmd.Flags().IntVarP(&height, "height", "H", 0, "target height")
	cmd.Flags().StringVar(&fit, "fit", "contain", "resize mode: contain, cover, or stretch")
	cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite an existing output file")
	return cmd
}
