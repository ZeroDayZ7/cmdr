package favicon

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	imageutil "github.com/zerodayz7/cmdr/internal/assets/image"
)

var defaultFaviconSizes = []int{16, 32, 48, 180, 192, 512}

func DefaultSizes() []int {
	out := make([]int, len(defaultFaviconSizes))
	copy(out, defaultFaviconSizes)
	return out
}

func GenerateFaviconSet(inputPath string, outputDir string, sizes []int, overwrite bool) ([]string, error) {
	return GenerateFaviconSetWithOptions(inputPath, outputDir, sizes, overwrite, true)
}

func GenerateFaviconSetWithOptions(inputPath string, outputDir string, sizes []int, overwrite bool, preserveTransparency bool) ([]string, error) {
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
		finalSizes = DefaultSizes()
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output directory %s: %w", outputDir, err)
	}

	stagingDir, err := os.MkdirTemp(outputDir, ".favicon-staging-")
	if err != nil {
		return nil, fmt.Errorf("create staging directory for favicons: %w", err)
	}
	defer os.RemoveAll(stagingDir)

	planned := make([]struct {
		size int
		path string
	}, 0, len(finalSizes))
	for _, size := range finalSizes {
		outputName := faviconFilename(size)
		outputPath := filepath.Join(outputDir, outputName)
		if info, err := os.Stat(outputPath); err == nil {
			if info.IsDir() {
				return nil, fmt.Errorf("output path is a directory: %s", outputPath)
			}
			if !overwrite {
				return nil, fmt.Errorf("output file already exists: %s", outputPath)
			}
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("check output file %s: %w", outputPath, err)
		}
		planned = append(planned, struct {
			size int
			path string
		}{size: size, path: outputPath})
	}

	generated := make([]string, 0, len(finalSizes))
	for _, item := range planned {
		stagingPath := filepath.Join(stagingDir, faviconFilename(item.size))
		resized, err := imageutil.ResizeWithFit(img, item.size, item.size, "contain")
		if err != nil {
			return nil, fmt.Errorf("resize favicon %dx%d: %w", item.size, item.size, err)
		}
		if !preserveTransparency {
			resized = imageutil.CompositeOnWhite(resized)
		}
		if err := imageutil.EncodeImage(stagingPath, resized, "png", true); err != nil {
			return nil, fmt.Errorf("write favicon %s: %w", stagingPath, err)
		}
		generated = append(generated, filepath.Base(stagingPath))
	}

	type backupEntry struct {
		backupPath string
		finalPath  string
	}
	backups := make([]backupEntry, 0, len(planned))
	created := make([]string, 0, len(planned))
	for _, item := range planned {
		finalPath := item.path
		stagingPath := filepath.Join(stagingDir, filepath.Base(finalPath))
		if info, err := os.Stat(finalPath); err == nil {
			if info.IsDir() {
				return nil, fmt.Errorf("output path is a directory: %s", finalPath)
			}
			backupPath := filepath.Join(stagingDir, ".backup-"+filepath.Base(finalPath))
			if err := copyFile(finalPath, backupPath); err != nil {
				return nil, fmt.Errorf("backup existing favicon %s: %w", finalPath, err)
			}
			backups = append(backups, backupEntry{backupPath: backupPath, finalPath: finalPath})
		} else if !os.IsNotExist(err) {
			return nil, fmt.Errorf("check output file %s: %w", finalPath, err)
		}
		if err := imageutil.ReplaceFile(stagingPath, finalPath, true); err != nil {
			for i := len(backups) - 1; i >= 0; i-- {
				entry := backups[i]
				if errRestore := copyFile(entry.backupPath, entry.finalPath); errRestore != nil {
					return nil, fmt.Errorf("commit favicon %s: %w (restore backup failed: %v)", finalPath, err, errRestore)
				}
			}
			for _, createdPath := range created {
				_ = os.Remove(createdPath)
			}
			return nil, fmt.Errorf("commit favicon %s: %w", finalPath, err)
		}
		if _, err := os.Stat(finalPath); err == nil {
			if _, backupErr := os.Stat(filepath.Join(stagingDir, ".backup-"+filepath.Base(finalPath))); backupErr == nil {
				continue
			}
			created = append(created, finalPath)
		}
	}
	return generated, nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
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
		if size > 2048 {
			return nil, fmt.Errorf("favicon size %d exceeds supported maximum 2048", size)
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
