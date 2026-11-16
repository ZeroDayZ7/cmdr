package favicon

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	imageutil "github.com/zerodayz7/cmdr/internal/assets/image"
)

var DefaultSizes = []int{16, 32, 48, 180, 192, 512}

func GenerateFaviconSet(inputPath string, outputDir string, sizes []int, overwrite bool) ([]string, error) {
	if strings.TrimSpace(inputPath) == "" {
		return nil, fmt.Errorf("input image is required")
	}
	if _, err := os.Stat(inputPath); err != nil {
		return nil, fmt.Errorf("source image %s does not exist: %w", inputPath, err)
	}

	img, _, err := imageutil.DecodeImage(inputPath)
	if err != nil {
		return nil, err
	}

	finalSizes, err := normalizeSizes(sizes)
	if err != nil {
		return nil, err
	}
	if len(finalSizes) == 0 {
		finalSizes = append([]int{}, DefaultSizes...)
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output directory %s: %w", outputDir, err)
	}

	generated := make([]string, 0, len(finalSizes))
	for _, size := range finalSizes {
		if size <= 0 {
			return nil, fmt.Errorf("favicon size must be positive, got %d", size)
		}
		outputName := faviconFilename(size)
		outputPath := filepath.Join(outputDir, outputName)
		if _, err := os.Stat(outputPath); err == nil && !overwrite {
			return nil, fmt.Errorf("output file already exists: %s", outputPath)
		}

		resized, err := imageutil.ResizeWithFit(img, size, size, "contain")
		if err != nil {
			return nil, fmt.Errorf("resize favicon %dx%d: %w", size, size, err)
		}
		if err := imageutil.EncodeImage(outputPath, resized, "png", true); err != nil {
			return nil, fmt.Errorf("write favicon %s: %w", outputPath, err)
		}
		generated = append(generated, outputName)
	}
	return generated, nil
}

func normalizeSizes(sizes []int) ([]int, error) {
	if len(sizes) == 0 {
		return nil, nil
	}
	seen := map[int]bool{}
	normalized := make([]int, 0, len(sizes))
	for _, size := range sizes {
		if size <= 0 {
			return nil, fmt.Errorf("favicon size must be positive, got %d", size)
		}
		if !seen[size] {
			seen[size] = true
			normalized = append(normalized, size)
		}
	}
	sort.Ints(normalized)
	return normalized, nil
}

func faviconFilename(size int) string {
	switch size {
	case 180:
		return "apple-touch-icon.png"
	case 192:
		return "icon-192.png"
	case 512:
		return "icon-512.png"
	default:
		return fmt.Sprintf("favicon-%dx%d.png", size, size)
	}
}
