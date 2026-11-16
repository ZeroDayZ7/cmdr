package imageutil

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	xdraw "golang.org/x/image/draw"
	"golang.org/x/sys/windows"
)

const (
	maxImageDimension = 16384
	maxPixelCount     = 268435456
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

	cfg, format, err := image.DecodeConfig(f)
	if err != nil {
		return nil, "", fmt.Errorf("invalid image file %s: %w", path, err)
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return nil, "", fmt.Errorf("image %s has invalid dimensions %dx%d", path, cfg.Width, cfg.Height)
	}
	if cfg.Width > maxImageDimension || cfg.Height > maxImageDimension {
		return nil, "", fmt.Errorf("image %s exceeds max allowed dimensions %dx%d", path, cfg.Width, cfg.Height)
	}
	if uint64(cfg.Width)*uint64(cfg.Height) > maxPixelCount {
		return nil, "", fmt.Errorf("image %s exceeds max pixel count %d", path, maxPixelCount)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, "", fmt.Errorf("reset image reader for %s: %w", path, err)
	}

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

func NormalizeOutputFormat(path, format string) (string, error) {
	if strings.TrimSpace(format) == "" {
		return OutputFormatFromPath(path)
	}
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "jpg" {
		format = "jpeg"
	}
	if !IsSupportedOutputFormat(format) {
		return "", fmt.Errorf("unsupported output format %q", format)
	}
	pathExt := strings.ToLower(filepath.Ext(path))
	if pathExt == ".png" && format != "png" {
		return "", fmt.Errorf("output path %s indicates PNG, but format %q was requested", path, format)
	}
	if (pathExt == ".jpg" || pathExt == ".jpeg") && format != "jpeg" {
		return "", fmt.Errorf("output path %s indicates JPEG, but format %q was requested", path, format)
	}
	if pathExt == ".gif" && format != "gif" {
		return "", fmt.Errorf("output path %s indicates GIF, but format %q was requested", path, format)
	}
	return format, nil
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
	if width, height := img.Bounds().Dx(), img.Bounds().Dy(); width > maxImageDimension || height > maxImageDimension {
		return fmt.Errorf("output image exceeds max allowed dimensions %dx%d", width, height)
	}

	resolvedFormat, err := NormalizeOutputFormat(path, format)
	if err != nil {
		return err
	}
	if info, err := os.Stat(path); err == nil && !info.Mode().IsRegular() {
		return fmt.Errorf("output path is not a regular file: %s", path)
	} else if err == nil && !overwrite {
		return fmt.Errorf("output file already exists: %s", path)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("check output file %s: %w", path, err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output directory for %s: %w", path, err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(path), "cmdr-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary output file for %s: %w", path, err)
	}
	tmpPath := tmpFile.Name()
	cleanup := func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpPath)
	}

	imgToEncode := img
	if resolvedFormat == "jpeg" && hasAlpha(img) {
		imgToEncode = compositeOnWhite(img)
	}

	if err := encodeImage(tmpFile, imgToEncode, resolvedFormat); err != nil {
		cleanup()
		return fmt.Errorf("encode image %s: %w", path, err)
	}
	if err := tmpFile.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temporary output file for %s: %w", path, err)
	}
	if err := ReplaceFile(tmpPath, path, overwrite); err != nil {
		_ = os.Remove(tmpPath)
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
	if srcW > maxImageDimension || srcH > maxImageDimension {
		return nil, fmt.Errorf("source image exceeds max allowed dimensions %dx%d", srcW, srcH)
	}

	fit = strings.ToLower(strings.TrimSpace(fit))
	if fit == "" {
		fit = "contain"
	}

	out := image.NewNRGBA(image.Rect(0, 0, targetW, targetH))
	if targetW > maxImageDimension || targetH > maxImageDimension {
		return nil, fmt.Errorf("target image exceeds max allowed dimensions %dx%d", targetW, targetH)
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
		xdraw.CatmullRom.Scale(out, image.Rect(offsetX, offsetY, offsetX+newW, offsetY+newH), img, img.Bounds(), draw.Over, nil)
		return out, nil
	case "cover":
		scale := math.Max(float64(targetW)/float64(srcW), float64(targetH)/float64(srcH))
		newW := int(math.Round(float64(srcW) * scale))
		newH := int(math.Round(float64(srcH) * scale))
		offsetX := (targetW - newW) / 2
		offsetY := (targetH - newH) / 2
		xdraw.CatmullRom.Scale(out, image.Rect(offsetX, offsetY, offsetX+newW, offsetY+newH), img, img.Bounds(), draw.Over, nil)
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported resize fit mode %q: expected contain, cover, or stretch", fit)
	}
}

func scaleImage(img image.Image, targetW, targetH int) image.Image {
	out := image.NewNRGBA(image.Rect(0, 0, targetW, targetH))
	xdraw.CatmullRom.Scale(out, out.Bounds(), img, img.Bounds(), draw.Over, nil)
	return out
}

func encodeImage(w io.Writer, img image.Image, format string) error {
	switch format {
	case "png":
		return png.Encode(w, img)
	case "jpeg":
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 90})
	case "gif":
		return gif.Encode(w, img, nil)
	default:
		return fmt.Errorf("unsupported output format %q", format)
	}
}

func hasAlpha(img image.Image) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if rgba := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA); rgba.A != 255 {
				return true
			}
		}
	}
	return false
}

func compositeOnWhite(img image.Image) image.Image {
	bounds := img.Bounds()
	out := image.NewRGBA(bounds)
	draw.Draw(out, bounds, &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(out, bounds, img, bounds.Min, draw.Over)
	return out
}

func ReplaceFile(src, dst string, overwrite bool) error {
	if src == "" || dst == "" {
		return errors.New("source and destination paths are required")
	}
	if _, err := os.Stat(dst); err == nil && !overwrite {
		return fmt.Errorf("file already exists: %s", dst)
	}
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := os.Remove(dst); err == nil || os.IsNotExist(err) {
		if err := os.Rename(src, dst); err == nil {
			return nil
		}
	}
	if os.Getenv("OS") == "Windows_NT" {
		srcPtr, _ := windows.UTF16PtrFromString(src)
		dstPtr, _ := windows.UTF16PtrFromString(dst)
		flags := uint32(windows.MOVEFILE_REPLACE_EXISTING | windows.MOVEFILE_WRITE_THROUGH)
		if err := windows.MoveFileEx(srcPtr, dstPtr, flags); err == nil {
			return nil
		}
	}
	return fmt.Errorf("replace file %s with %s", src, dst)
}
