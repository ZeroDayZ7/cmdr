package convert

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	imageutil "github.com/zerodayz7/cmdr/internal/assets/image"
)

func ConvertImage(srcPath, dstPath, requestedFormat string, quality int, lossless, overwrite bool) (string, error) {
	if strings.TrimSpace(srcPath) == "" {
		return "", fmt.Errorf("input image path is required")
	}
	if strings.TrimSpace(dstPath) == "" {
		return "", fmt.Errorf("output image path is required")
	}

	samePath, err := sameFilePath(srcPath, dstPath)
	if err != nil {
		return "", fmt.Errorf("validate source and target paths: %w", err)
	}
	if samePath {
		return "", fmt.Errorf("input and output paths must differ: %s", srcPath)
	}

	format, err := imageutil.NormalizeOutputFormat(dstPath, requestedFormat)
	if err != nil {
		return "", err
	}
	if format != "webp" {
		if quality != 85 {
			return "", fmt.Errorf("--quality is supported only for WebP output")
		}
		if lossless {
			return "", fmt.Errorf("--lossless is supported only for WebP output")
		}
	}
	if format == "webp" && (quality < 1 || quality > 100) {
		return "", fmt.Errorf("WebP quality must be between 1 and 100: got %d", quality)
	}

	img, _, err := imageutil.DecodeImage(srcPath)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return "", fmt.Errorf("create output directory for %s: %w", dstPath, err)
	}
	options := imageutil.EncodeOptions{Quality: quality, Lossless: lossless}
	if err := imageutil.EncodeImageWithOptions(dstPath, img, format, overwrite, options); err != nil {
		return "", err
	}
	return format, nil
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
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if resolvedB, err := filepath.EvalSymlinks(baseB); err == nil {
		baseB = resolvedB
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if strings.EqualFold(filepath.Clean(baseA), filepath.Clean(baseB)) {
		return true, nil
	}
	infoA, errA := os.Stat(baseA)
	infoB, errB := os.Stat(baseB)
	if errA == nil && errB == nil {
		return os.SameFile(infoA, infoB), nil
	}
	return false, nil
}
