package resize

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	imageutil "github.com/zerodayz7/cmdr/internal/assets/image"
)

func ResizeImage(srcPath, dstPath string, width, height int, fit string, overwrite bool) error {
	if srcPath == "" {
		return fmt.Errorf("input image path is required")
	}
	if dstPath == "" {
		return fmt.Errorf("output image path is required")
	}
	if width <= 0 || height <= 0 {
		return fmt.Errorf("target dimensions must be positive, got %dx%d", width, height)
	}
	if strings.TrimSpace(fit) == "" {
		fit = "contain"
	}

	img, _, err := imageutil.DecodeImage(srcPath)
	if err != nil {
		return err
	}

	if filepath.Clean(srcPath) == filepath.Clean(dstPath) {
		return fmt.Errorf("input and output paths must differ: %s", srcPath)
	}
	if _, err := os.Stat(dstPath); err == nil && !overwrite {
		return fmt.Errorf("output file already exists: %s", dstPath)
	}

	resized, err := imageutil.ResizeWithFit(img, width, height, fit)
	if err != nil {
		return err
	}

	format, err := imageutil.OutputFormatFromPath(dstPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return fmt.Errorf("create output directory for %s: %w", dstPath, err)
	}
	return imageutil.EncodeImage(dstPath, resized, format, true)
}
