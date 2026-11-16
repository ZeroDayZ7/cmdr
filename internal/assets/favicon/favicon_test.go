package favicon

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func writeTestImage(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create image: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("encode image: %v", err)
	}
}

func TestGenerateFaviconSetProducesAllSizes(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "public")
	src := filepath.Join(t.TempDir(), "logo.png")
	img := image.NewNRGBA(image.Rect(0, 0, 512, 256))
	for y := 0; y < 256; y++ {
		for x := 0; x < 512; x++ {
			if x < 256 {
				img.SetNRGBA(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
			} else {
				img.SetNRGBA(x, y, color.NRGBA{R: 0, G: 255, B: 0, A: 255})
			}
		}
	}
	writeTestImage(t, src, img)

	files, err := GenerateFaviconSet(src, outDir, []int{16, 32, 48, 180, 192, 512}, false)
	if err != nil {
		t.Fatalf("GenerateFaviconSet returned error: %v", err)
	}
	if len(files) != 6 {
		t.Fatalf("expected 6 files, got %d: %v", len(files), files)
	}
	for _, name := range []string{"favicon-16x16.png", "favicon-32x32.png", "favicon-48x48.png", "apple-touch-icon.png", "icon-192.png", "icon-512.png"} {
		if _, err := os.Stat(filepath.Join(outDir, name)); err != nil {
			t.Fatalf("expected file %s to exist: %v", name, err)
		}
	}
}

func TestGenerateFaviconSetPreservesTransparency(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "public")
	src := filepath.Join(t.TempDir(), "logo.png")
	img := image.NewNRGBA(image.Rect(0, 0, 128, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 128; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	writeTestImage(t, src, img)

	_, err := GenerateFaviconSet(src, outDir, []int{16}, false)
	if err != nil {
		t.Fatalf("GenerateFaviconSet returned error: %v", err)
	}
	path := filepath.Join(outDir, "favicon-16x16.png")
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open generated favicon: %v", err)
	}
	defer f.Close()
	decoded, err := png.Decode(f)
	if err != nil {
		t.Fatalf("decode generated favicon: %v", err)
	}
	if rgba := color.NRGBAModel.Convert(decoded.At(0, 0)).(color.NRGBA); rgba.A == 255 {
		t.Fatal("expected transparent padding around the scaled image")
	}
}

func TestGenerateFaviconSetRejectsInvalidInput(t *testing.T) {
	if _, err := GenerateFaviconSet(filepath.Join(t.TempDir(), "missing.png"), t.TempDir(), []int{16}, false); err == nil {
		t.Fatal("expected missing source image to error")
	}
}

func TestGenerateFaviconSetRejectsInvalidSizes(t *testing.T) {
	src := filepath.Join(t.TempDir(), "logo.png")
	writeTestImage(t, src, image.NewNRGBA(image.Rect(0, 0, 32, 32)))
	if _, err := GenerateFaviconSet(src, t.TempDir(), []int{0, -2}, false); err == nil {
		t.Fatal("expected invalid favicon sizes to error")
	}
}

func TestGenerateFaviconSetRefusesExistingOutputWithoutOverwrite(t *testing.T) {
	src := filepath.Join(t.TempDir(), "logo.png")
	writeTestImage(t, src, image.NewNRGBA(image.Rect(0, 0, 32, 32)))
	outDir := filepath.Join(t.TempDir(), "public")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("create output dir: %v", err)
	}
	outputFile := filepath.Join(outDir, "favicon-16x16.png")
	if err := os.WriteFile(outputFile, []byte("existing"), 0o644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}
	if _, err := GenerateFaviconSet(src, outDir, []int{16}, false); err == nil {
		t.Fatal("expected existing output file to be refused")
	}
}

func TestGenerateFaviconSetAllowsOverwrite(t *testing.T) {
	src := filepath.Join(t.TempDir(), "logo.png")
	writeTestImage(t, src, image.NewNRGBA(image.Rect(0, 0, 32, 32)))
	outDir := filepath.Join(t.TempDir(), "public")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("create output dir: %v", err)
	}
	outputFile := filepath.Join(outDir, "favicon-16x16.png")
	if err := os.WriteFile(outputFile, []byte("existing"), 0o644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}
	if _, err := GenerateFaviconSet(src, outDir, []int{16}, true); err != nil {
		t.Fatalf("overwrite expected to succeed: %v", err)
	}
}
