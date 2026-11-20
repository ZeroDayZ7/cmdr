package convert

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	imageutil "github.com/zerodayz7/cmdr/internal/assets/image"
)

func writeTestImage(t *testing.T, path, format string, transparent bool) {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 32, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 32; x++ {
			alpha := uint8(255)
			if transparent && x < 16 {
				alpha = 96
			}
			img.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 7), G: uint8(y * 9), B: 180, A: alpha})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	switch format {
	case "png":
		err = png.Encode(f, img)
	case "jpeg":
		err = jpeg.Encode(f, img, &jpeg.Options{Quality: 90})
	default:
		t.Fatalf("unsupported test format %q", format)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestConvertCommonPairsAndReopen(t *testing.T) {
	pairs := [][2]string{{"png", "webp"}, {"jpeg", "webp"}, {"webp", "png"}, {"webp", "jpeg"}, {"webp", "webp"}}
	for _, pair := range pairs {
		t.Run(pair[0]+"-"+pair[1], func(t *testing.T) {
			dir := t.TempDir()
			srcFormat := pair[0]
			src := filepath.Join(dir, "source."+srcFormat)
			if srcFormat == "jpeg" {
				writeTestImage(t, src, "jpeg", false)
			} else if srcFormat == "png" {
				writeTestImage(t, src, "png", true)
			} else {
				writeTestImage(t, filepath.Join(dir, "seed.png"), "png", true)
				if _, err := ConvertImage(filepath.Join(dir, "seed.png"), src, "webp", 85, false, false); err != nil {
					t.Fatal(err)
				}
			}
			dst := filepath.Join(dir, "output."+pair[1])
			resolved, err := ConvertImage(src, dst, "", 85, false, false)
			if err != nil {
				t.Fatalf("ConvertImage: %v", err)
			}
			if resolved != pair[1] && !(pair[1] == "jpeg" && resolved == "jpeg") {
				t.Fatalf("format = %q", resolved)
			}
			img, gotFormat, err := imageutil.DecodeImage(dst)
			if err != nil {
				t.Fatalf("reopen output: %v", err)
			}
			wantFormat := pair[1]
			if gotFormat == "jpeg" && wantFormat == "jpeg" {
				wantFormat = "jpeg"
			}
			if gotFormat != wantFormat {
				t.Fatalf("decoded format = %q, want %q", gotFormat, wantFormat)
			}
			if img.Bounds().Dx() != 32 || img.Bounds().Dy() != 24 {
				t.Fatalf("dimensions = %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
			}
		})
	}
}

func TestConvertPreservesPNGAlphaThroughWebP(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.png")
	webpPath := filepath.Join(dir, "output.webp")
	pngPath := filepath.Join(dir, "roundtrip.png")
	writeTestImage(t, src, "png", true)
	if _, err := ConvertImage(src, webpPath, "", 85, false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := ConvertImage(webpPath, pngPath, "", 85, false, false); err != nil {
		t.Fatal(err)
	}
	decoded, _, err := imageutil.DecodeImage(pngPath)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, alpha := decoded.At(0, 12).RGBA()
	if alpha == 0xffff {
		t.Fatal("transparent pixel became fully opaque")
	}
}

func TestConvertLosslessAndLossy(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.png")
	writeTestImage(t, src, "png", true)
	for _, tc := range []struct {
		name     string
		quality  int
		lossless bool
	}{
		{"lossy", 20, false}, {"lossless", 100, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dst := filepath.Join(dir, tc.name+".webp")
			if _, err := ConvertImage(src, dst, "", tc.quality, tc.lossless, false); err != nil {
				t.Fatal(err)
			}
			if _, _, err := imageutil.DecodeImage(dst); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestConvertValidationAndSafety(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.png")
	writeTestImage(t, src, "png", false)

	if _, err := ConvertImage(src, filepath.Join(dir, "bad.bin"), "", 85, false, false); err == nil {
		t.Fatal("expected unsupported extension error")
	}
	if _, err := ConvertImage(src, filepath.Join(dir, "output.png"), "webp", 85, false, false); err == nil {
		t.Fatal("expected format mismatch error")
	}
	if _, err := ConvertImage(filepath.Join(dir, "missing.png"), filepath.Join(dir, "out.webp"), "", 85, false, false); err == nil {
		t.Fatal("expected missing input error")
	}
	if _, err := ConvertImage(src, src, "", 85, false, false); err == nil {
		t.Fatal("expected same-path error")
	}
	if _, err := ConvertImage(src, filepath.Join(dir, "out.webp"), "", 101, false, false); err == nil {
		t.Fatal("expected quality range error")
	}

	dst := filepath.Join(dir, "existing.webp")
	if err := os.WriteFile(dst, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ConvertImage(src, dst, "", 85, false, false); err == nil {
		t.Fatal("expected existing output error")
	}
	data, _ := os.ReadFile(dst)
	if string(data) != "old" {
		t.Fatal("existing output changed without overwrite")
	}
	if _, err := ConvertImage(src, dst, "", 85, false, true); err != nil {
		t.Fatalf("overwrite: %v", err)
	}
	if _, _, err := imageutil.DecodeImage(dst); err != nil {
		t.Fatalf("overwrite output was not decodable: %v", err)
	}

	corrupt := filepath.Join(dir, "corrupt.png")
	if err := os.WriteFile(corrupt, []byte("not an image"), 0o644); err != nil {
		t.Fatal(err)
	}
	corruptDst := filepath.Join(dir, "corrupt.webp")
	if _, err := ConvertImage(corrupt, corruptDst, "", 85, false, false); err == nil {
		t.Fatal("expected corrupt input error")
	}
	entries, _ := os.ReadDir(dir)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "cmdr-") && strings.HasSuffix(entry.Name(), ".tmp") {
			t.Fatalf("temporary file left after failure: %s", entry.Name())
		}
	}
}
