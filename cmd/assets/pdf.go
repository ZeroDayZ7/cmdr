package assets

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	assetpdf "github.com/zerodayz7/cmdr/internal/assets/pdf"
)

func NewPDFCmd() *cobra.Command {
	var outputName, pageSize, orientation, preset, maxSize string
	var margin float64
	var force, recursive bool

	cmd := &cobra.Command{
		Use:   "pdf <directory>",
		Short: "Create a PDF document from numbered images",
		Long:  "Create an A4 or Letter PDF from PNG, JPG/JPEG, and WebP files in a directory. Images must start with a numeric prefix.",
		Example: `  cmdr assets pdf ./pages
  cmdr assets pdf ./pages -o comic-001.pdf
  cmdr assets pdf ./pages --preset ai --orientation portrait
  cmdr assets pdf ./pages --recursive --max-size 2MB`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			directory := args[0]
			files, unnumbered, err := assetpdf.CollectImages(directory, recursive)
			if err != nil {
				return err
			}
			normalizedPreset, err := assetpdf.NormalizePreset(preset)
			if err != nil {
				return err
			}
			maxSizeBytes, err := assetpdf.ParseSize(maxSize)
			if err != nil {
				return err
			}
			profile, _ := assetpdf.GetProfile(normalizedPreset)

			cmd.Printf("CMDR Assets PDF\n\nSource: %s\n", directory)
			cmd.Printf("Images found: %d\n", len(files))
			cmd.Printf("Formats: PNG: %d | WebP: %d | JPG: %d\n", countFormat(files, "PNG"), countFormat(files, "WebP"), countFormat(files, "JPG"))
			cmd.Printf("Preset: %s\n", profile.Name)
			cmd.Printf("Page size: %s\n", strings.ToUpper(defaultIfEmpty(pageSize, "A4")))
			cmd.Printf("Orientation: %s\n", titleCase(defaultIfEmpty(orientation, "auto")))
			cmd.Printf("Margins: %g mm\nSorting: Numeric\n", margin)
			if maxSizeBytes > 0 {
				cmd.Printf("Target size: %s\n", maxSize)
			}
			for _, name := range unnumbered {
				cmd.Printf("[SKIP] %s (missing numeric prefix)\n", name)
			}
			for _, file := range files {
				cmd.Printf("[OK] %s\n", file.Name)
			}

			result, err := assetpdf.Generate(directory, assetpdf.Options{
				OutputName:   outputName,
				PageSize:     pageSize,
				Orientation:  orientation,
				MarginMM:     margin,
				Recursive:    recursive,
				Overwrite:    force,
				Preset:       normalizedPreset,
				MaxSizeBytes: maxSizeBytes,
			})
			if err != nil {
				return err
			}
			cmd.Printf("\nOptimization:\n")
			cmd.Printf("Original images: %s\n", formatBytes(result.OriginalImageBytes))
			cmd.Printf("Optimized images: %s\n", formatBytes(result.OptimizedImageBytes))
			cmd.Printf("\nPDF:\nGenerated PDF: %s\n", formatBytes(result.PDFBytes))
			if result.OriginalImageBytes > 0 {
				cmd.Printf("Image size reduction: %.1f%%\n", reductionPercent(result.OriginalImageBytes, result.OptimizedImageBytes))
			}
			if maxSizeBytes > 0 && !result.SizeLimitReached {
				cmd.Printf("[WARNING] Target size was not reached; retained quality level: %d DPI, JPEG %d\n", result.QualityDPI, result.JPEGQuality)
			}
			cmd.Printf("\n[SUCCESS] PDF created successfully.\n\nOutput: %s\nPages: %d\n", result.OutputPath, len(result.Files))
			return nil
		},
	}

	cmd.Flags().StringVarP(&outputName, "output", "o", "", "output PDF file name or path")
	cmd.Flags().StringVar(&pageSize, "page-size", "A4", "page size: A4 or LETTER")
	cmd.Flags().StringVar(&orientation, "orientation", "auto", "orientation: auto, portrait, or landscape")
	cmd.Flags().Float64Var(&margin, "margin", 0, "page margin in millimeters")
	cmd.Flags().BoolVar(&force, "force", false, "overwrite an existing output PDF")
	cmd.Flags().BoolVar(&recursive, "recursive", false, "search subdirectories recursively")
	cmd.Flags().StringVar(&preset, "preset", "print", "quality preset: print, balanced, ai, or preview")
	cmd.Flags().StringVar(&maxSize, "max-size", "", "maximum PDF size, for example 800KB or 2MB")
	return cmd
}

func countFormat(files []assetpdf.ImageFile, format string) int {
	count := 0
	for _, file := range files {
		if file.Format == format {
			count++
		}
	}
	return count
}

func defaultIfEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func titleCase(value string) string {
	value = strings.ToLower(value)
	if value == "" {
		return ""
	}
	return fmt.Sprintf("%c%s", value[0]-32, value[1:])
}

func formatBytes(value int64) string {
	if value < 1000 {
		return fmt.Sprintf("%d B", value)
	}
	if value < 1000*1000 {
		return fmt.Sprintf("%.1f KB", float64(value)/1000)
	}
	return fmt.Sprintf("%.1f MB", float64(value)/(1000*1000))
}

func reductionPercent(original, optimized int64) float64 {
	if original <= 0 {
		return 0
	}
	return (1 - float64(optimized)/float64(original)) * 100
}
