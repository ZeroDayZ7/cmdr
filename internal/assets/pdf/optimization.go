package pdf

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math/big"
	"os"
	"regexp"
	"strings"

	xdraw "golang.org/x/image/draw"
)

var sizePattern = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)([A-Z]*|B)$`)

type QualityProfile struct {
	Name   string
	Levels []QualityLevel
}

type QualityLevel struct {
	DPI         int
	JPEGQuality int
}

var qualityProfiles = map[string]QualityProfile{
	"print":    {Name: "Print", Levels: []QualityLevel{{DPI: 0, JPEGQuality: 100}}},
	"balanced": {Name: "Balanced", Levels: []QualityLevel{{DPI: 200, JPEGQuality: 82}, {DPI: 180, JPEGQuality: 75}, {DPI: 160, JPEGQuality: 68}}},
	"ai":       {Name: "AI", Levels: []QualityLevel{{DPI: 144, JPEGQuality: 70}, {DPI: 132, JPEGQuality: 68}, {DPI: 120, JPEGQuality: 65}, {DPI: 108, JPEGQuality: 62}, {DPI: 96, JPEGQuality: 58}}},
	"preview":  {Name: "Preview", Levels: []QualityLevel{{DPI: 96, JPEGQuality: 55}, {DPI: 80, JPEGQuality: 50}, {DPI: 72, JPEGQuality: 45}}},
}

func GetProfile(preset string) (QualityProfile, bool) {
	profile, ok := qualityProfiles[preset]
	if !ok {
		return QualityProfile{}, false
	}
	profile.Levels = append([]QualityLevel(nil), profile.Levels...)
	return profile, true
}

func NormalizePreset(value string) (string, error) {
	preset := strings.ToLower(strings.TrimSpace(value))
	if preset == "" {
		preset = "print"
	}
	if _, ok := GetProfile(preset); !ok {
		return "", fmt.Errorf("invalid preset %q: expected print, balanced, ai, or preview", value)
	}
	return preset, nil
}

func ParseSize(value string) (int64, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return 0, nil
	}
	match := sizePattern.FindStringSubmatch(value)
	if len(match) != 3 {
		return 0, fmt.Errorf("invalid max size %q: use values like 800KB or 2MB", value)
	}
	amount, ok := new(big.Rat).SetString(match[1])
	if !ok || amount.Sign() <= 0 {
		return 0, fmt.Errorf("max size must be positive")
	}
	var multiplier int64
	switch match[2] {
	case "", "B":
		multiplier = 1
	case "KB":
		multiplier = 1000
	case "MB":
		multiplier = 1000 * 1000
	case "KIB":
		multiplier = 1024
	case "MIB":
		multiplier = 1024 * 1024
	default:
		return 0, fmt.Errorf("invalid max-size unit %q: use B, KB, MB, KIB, or MIB", match[2])
	}
	amount.Mul(amount, new(big.Rat).SetInt64(multiplier))
	if !amount.IsInt() {
		return 0, fmt.Errorf("max size is too precise")
	}
	bytesValue := amount.Num()
	if !bytesValue.IsInt64() || bytesValue.Sign() <= 0 {
		return 0, fmt.Errorf("max size exceeds supported range")
	}
	return bytesValue.Int64(), nil
}

func resizeToDPI(img image.Image, dpi int, pageSize fpdfSize) (image.Image, int, int) {
	if dpi <= 0 {
		bounds := img.Bounds()
		return img, bounds.Dx(), bounds.Dy()
	}
	pageW := int(float64(pageSize.w) / 25.4 * float64(dpi))
	pageH := int(float64(pageSize.h) / 25.4 * float64(dpi))
	bounds := img.Bounds()
	srcW, srcH := bounds.Dx(), bounds.Dy()
	if srcW <= pageW && srcH <= pageH {
		return img, srcW, srcH
	}
	scale := float64(pageW) / float64(srcW)
	if float64(pageH)/float64(srcH) < scale {
		scale = float64(pageH) / float64(srcH)
	}
	dstW := max(1, int(float64(srcW)*scale+0.5))
	dstH := max(1, int(float64(srcH)*scale+0.5))
	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	return dst, dstW, dstH
}

type fpdfSize struct{ w, h float64 }

type OptimizedImage struct {
	Data   []byte
	Format string
	Width  int
	Height int
}

func optimizeImage(img image.Image, preset string, level QualityLevel, size fpdfSize) ([]byte, string, error) {
	optimized, err := optimizeImageWithSize(img, preset, level, size)
	if err != nil {
		return nil, "", err
	}
	return optimized.Data, optimized.Format, nil
}

func optimizeImageWithSize(img image.Image, preset string, level QualityLevel, size fpdfSize) (OptimizedImage, error) {
	resized, width, height := resizeToDPI(img, level.DPI, size)
	var out bytes.Buffer
	if preset == "print" {
		if err := encodePNG(&out, resized); err != nil {
			return OptimizedImage{}, err
		}
		return OptimizedImage{Data: out.Bytes(), Format: "png", Width: width, Height: height}, nil
	}
	bounds := resized.Bounds()
	width, height = bounds.Dx(), bounds.Dy()
	white := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(white, white.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(white, white.Bounds(), resized, bounds.Min, draw.Over)
	if err := jpeg.Encode(&out, white, &jpeg.Options{Quality: level.JPEGQuality}); err != nil {
		return OptimizedImage{}, err
	}
	return OptimizedImage{Data: out.Bytes(), Format: "jpg", Width: width, Height: height}, nil
}

func encodePNG(out *bytes.Buffer, img image.Image) error {
	return png.Encode(out, img)
}

func originalImageBytes(files []ImageFile) (int64, error) {
	var total int64
	for _, file := range files {
		info, err := os.Stat(file.Path)
		if err != nil {
			return 0, err
		}
		total += info.Size()
	}
	return total, nil
}

func normalizePresetForFallback(value string) string {
	preset, err := NormalizePreset(value)
	if err != nil {
		return "print"
	}
	return preset
}
