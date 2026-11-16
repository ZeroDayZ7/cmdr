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

	samePath, err := sameFilePath(srcPath, dstPath)
	if err != nil {
		return fmt.Errorf("validate source and target paths: %w", err)
	}
	if samePath {
		return fmt.Errorf("input and output paths must differ: %s", srcPath)
	}

	img, _, err := imageutil.DecodeImage(srcPath)
	if err != nil {
		return err
	}

	if _, err := os.Stat(dstPath); err == nil && !overwrite {
		return fmt.Errorf("output file already exists: %s", dstPath)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("check output file %s: %w", dstPath, err)
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
	return imageutil.EncodeImage(dstPath, resized, format, overwrite)
}

func sameFilePath(pathA, pathB string) (bool, error) {
	baseA, err := filepath.Abs(pathA)
	if err != nil {
		return false, err
	}
	baseB, err := filepath.Abs(pathB)
	if err != nil {
		return false, err
	}

	if resolvedA, err := filepath.EvalSymlinks(baseA); err == nil {
		baseA = resolvedA
	}
	if resolvedB, err := filepath.EvalSymlinks(baseB); err == nil {
		baseB = resolvedB
	}

	if strings.EqualFold(baseA, baseB) {
		return true, nil
	}

	infoA, errA := os.Stat(baseA)
	infoB, errB := os.Stat(baseB)
	if errA == nil && errB == nil {
		return os.SameFile(infoA, infoB), nil
	}
	return false, nil
}
