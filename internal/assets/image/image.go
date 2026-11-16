package imageutil

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

func DecodeImage(path string) (image.Image, string, error) {
	if path == "" {
		return nil, "", errors.New("image path is empty")
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, "", fmt.Errorf("stat source image %s: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, "", fmt.Errorf("source image is not a regular file: %s", path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("open source image %s: %w", path, err)
	}
	defer f.Close()

	img, format, err := image.Decode(f)
	if err != nil {
		return nil, "", fmt.Errorf("invalid image file %s: %w", path, err)
	}
	if img == nil {
		return nil, "", fmt.Errorf("decoded image is empty: %s", path)
	}
	return img, strings.ToLower(format), nil
}

func IsSupportedOutputFormat(ext string) bool {
	switch strings.ToLower(strings.TrimPrefix(ext, ".")) {
	case "png", "jpg", "jpeg", "gif":
		return true
	default:
		return false
	}
}

func OutputFormatFromPath(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return "png", nil
	case ".jpg", ".jpeg":
		return "jpeg", nil
	case ".gif":
		return "gif", nil
	default:
		return "", fmt.Errorf("unsupported output format: %s", path)
	}
}

func EncodeImage(path string, img image.Image, format string, overwrite bool) error {
	if path == "" {
		return errors.New("output path is empty")
	}
	if img == nil {
		return errors.New("image is nil")
	}
	if width, height := img.Bounds().Dx(), img.Bounds().Dy(); width <= 0 || height <= 0 {
		return fmt.Errorf("image has invalid dimensions %dx%d", width, height)
	}
	if format == "" {
		resolved, err := OutputFormatFromPath(path)
		if err != nil {
			return err
		}
		format = resolved
	}
	if !IsSupportedOutputFormat(format) {
		return fmt.Errorf("unsupported output format %q", format)
	}
	if info, err := os.Stat(path); err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("output path is not a regular file: %s", path)
	} else if err == nil && !overwrite {
		return fmt.Errorf("output file already exists: %s", path)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output directory for %s: %w", path, err)
	}

	if _, err := os.Stat(path); err == nil && overwrite {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("remove existing file %s before overwrite: %w", path, err)
		}
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(path), "*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary output file for %s: %w", path, err)
	}
	cleanupPath := tmpFile.Name()
	defer os.Remove(cleanupPath)

	if err := func() error {
		defer tmpFile.Close()
		switch format {
		case "png":
			return png.Encode(tmpFile, img)
		case "jpg", "jpeg":
			return jpeg.Encode(tmpFile, img, &jpeg.Options{Quality: 90})
		case "gif":
			return gif.Encode(tmpFile, img, nil)
		default:
			return fmt.Errorf("unsupported output format %q", format)
		}
	}(); err != nil {
		return fmt.Errorf("encode image %s: %w", path, err)
	}

	if err := os.Rename(cleanupPath, path); err != nil {
		return fmt.Errorf("save output image %s: %w", path, err)
	}
	return nil
}

func ResizeWithFit(img image.Image, targetW, targetH int, fit string) (image.Image, error) {
	if img == nil {
		return nil, errors.New("image is nil")
	}
	if targetW <= 0 || targetH <= 0 {
		return nil, fmt.Errorf("target dimensions must be positive, got %dx%d", targetW, targetH)
	}

	srcW := img.Bounds().Dx()
	srcH := img.Bounds().Dy()
	if srcW <= 0 || srcH <= 0 {
		return nil, fmt.Errorf("source image has invalid dimensions %dx%d", srcW, srcH)
	}

	fit = strings.ToLower(strings.TrimSpace(fit))
	if fit == "" {
		fit = "contain"
	}

	out := image.NewNRGBA(image.Rect(0, 0, targetW, targetH))
	for x := 0; x < targetW; x++ {
		for y := 0; y < targetH; y++ {
			out.Set(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 0})
		}
	}

	switch fit {
	case "stretch":
		return scaleImage(img, targetW, targetH), nil
	case "contain":
		scale := math.Min(float64(targetW)/float64(srcW), float64(targetH)/float64(srcH))
		newW := int(math.Round(float64(srcW) * scale))
		newH := int(math.Round(float64(srcH) * scale))
		offsetX := (targetW - newW) / 2
		offsetY := (targetH - newH) / 2
		draw.CatmullRom.Scale(out, image.Rect(offsetX, offsetY, offsetX+newW, offsetY+newH), img, img.Bounds(), draw.Over, nil)
		return out, nil
	case "cover":
		scale := math.Max(float64(targetW)/float64(srcW), float64(targetH)/float64(srcH))
		newW := int(math.Round(float64(srcW) * scale))
		newH := int(math.Round(float64(srcH) * scale))
		offsetX := (targetW - newW) / 2
		offsetY := (targetH - newH) / 2
		draw.CatmullRom.Scale(out, image.Rect(offsetX, offsetY, offsetX+newW, offsetY+newH), img, img.Bounds(), draw.Over, nil)
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported resize fit mode %q: expected contain, cover, or stretch", fit)
	}
}

func scaleImage(img image.Image, targetW, targetH int) image.Image {
	out := image.NewNRGBA(image.Rect(0, 0, targetW, targetH))
	draw.CatmullRom.Scale(out, out.Bounds(), img, img.Bounds(), draw.Over, nil)
	return out
}
