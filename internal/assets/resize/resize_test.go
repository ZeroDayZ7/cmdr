package resize

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func writeImageFile(t *testing.T, path string, img image.Image, format string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create image: %v", err)
	}
	defer func() { _ = f.Close() }()

	switch format {
	case "png":
		if err := png.Encode(f, img); err != nil {
			t.Fatalf("encode png: %v", err)
		}
	case "jpeg":
		if err := jpeg.Encode(f, img, nil); err != nil {
			t.Fatalf("encode jpeg: %v", err)
		}
	default:
		t.Fatalf("unsupported test format: %s", format)
	}
}

func readImageDimensions(t *testing.T, path string) (int, int) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open output image: %v", err)
	}
	defer func() { _ = f.Close() }()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("decode output image: %v", err)
	}
	b := img.Bounds()
	return b.Dx(), b.Dy()
}

func TestResizeContain(t *testing.T) {
	src := filepath.Join(t.TempDir(), "source.png")
	img := image.NewNRGBA(image.Rect(0, 0, 1000, 500))
	for y := 0; y < 500; y++ {
		for x := 0; x < 1000; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	writeImageFile(t, src, img, "png")

	dst := filepath.Join(t.TempDir(), "out", "result.png")
	if err := ResizeImage(src, dst, 1200, 630, "contain", false); err != nil {
		t.Fatalf("ResizeImage returned error: %v", err)
	}
	w, h := readImageDimensions(t, dst)
	if w != 1200 || h != 630 {
		t.Fatalf("contain should keep the target canvas size at 1200x630, got %dx%d", w, h)
	}
}

func TestResizeCover(t *testing.T) {
	src := filepath.Join(t.TempDir(), "source.png")
	img := image.NewNRGBA(image.Rect(0, 0, 1000, 500))
	writeImageFile(t, src, img, "png")

	dst := filepath.Join(t.TempDir(), "out", "result.png")
	if err := ResizeImage(src, dst, 1200, 630, "cover", false); err != nil {
		t.Fatalf("ResizeImage returned error: %v", err)
	}
	w, h := readImageDimensions(t, dst)
	if w != 1200 || h != 630 {
		t.Fatalf("cover should resize to 1200x630, got %dx%d", w, h)
	}
}

func TestResizeStretch(t *testing.T) {
	src := filepath.Join(t.TempDir(), "source.png")
	img := image.NewNRGBA(image.Rect(0, 0, 1000, 500))
	writeImageFile(t, src, img, "png")

	dst := filepath.Join(t.TempDir(), "out", "result.png")
	if err := ResizeImage(src, dst, 1200, 630, "stretch", false); err != nil {
		t.Fatalf("ResizeImage returned error: %v", err)
	}
	w, h := readImageDimensions(t, dst)
	if w != 1200 || h != 630 {
		t.Fatalf("stretch should resize to 1200x630, got %dx%d", w, h)
	}
}

func TestResizeSquareImage(t *testing.T) {
	src := filepath.Join(t.TempDir(), "source.png")
	img := image.NewNRGBA(image.Rect(0, 0, 512, 512))
	writeImageFile(t, src, img, "png")

	dst := filepath.Join(t.TempDir(), "out", "result.jpg")
	if err := ResizeImage(src, dst, 256, 256, "contain", false); err != nil {
		t.Fatalf("ResizeImage returned error: %v", err)
	}
	w, h := readImageDimensions(t, dst)
	if w != 256 || h != 256 {
		t.Fatalf("square image should resize to 256x256, got %dx%d", w, h)
	}
}

func TestResizeHorizontalAndVerticalImages(t *testing.T) {
	for _, tc := range []struct {
		name  string
		size  image.Rectangle
		wantW int
		wantH int
		fit   string
	}{
		{name: "horizontal", size: image.Rect(0, 0, 2000, 1000), wantW: 1200, wantH: 600, fit: "contain"},
		{name: "vertical", size: image.Rect(0, 0, 1000, 2000), wantW: 1200, wantH: 600, fit: "contain"},
	} {
		src := filepath.Join(t.TempDir(), tc.name+".png")
		img := image.NewNRGBA(tc.size)
		writeImageFile(t, src, img, "png")
		dst := filepath.Join(t.TempDir(), "out", tc.name+".png")
		if err := ResizeImage(src, dst, 1200, 600, tc.fit, false); err != nil {
			t.Fatalf("ResizeImage returned error for %s: %v", tc.name, err)
		}
		w, h := readImageDimensions(t, dst)
		if w != tc.wantW || h != tc.wantH {
			t.Fatalf("%s expected %dx%d but got %dx%d", tc.name, tc.wantW, tc.wantH, w, h)
		}
	}
}

func TestResizeRejectsInvalidDimensions(t *testing.T) {
	src := filepath.Join(t.TempDir(), "source.png")
	writeImageFile(t, src, image.NewNRGBA(image.Rect(0, 0, 32, 32)), "png")
	if err := ResizeImage(src, filepath.Join(t.TempDir(), "out.png"), 0, 100, "contain", false); err == nil {
		t.Fatal("expected invalid width to error")
	}
	if err := ResizeImage(src, filepath.Join(t.TempDir(), "out.png"), 100, 0, "contain", false); err == nil {
		t.Fatal("expected invalid height to error")
	}
}

func TestResizeSupportsWebPFormat(t *testing.T) {
	src := filepath.Join(t.TempDir(), "source.png")
	writeImageFile(t, src, image.NewNRGBA(image.Rect(0, 0, 32, 32)), "png")
	dst := filepath.Join(t.TempDir(), "out.webp")
	if err := ResizeImage(src, dst, 100, 100, "contain", true); err != nil {
		t.Fatalf("ResizeImage WebP returned error: %v", err)
	}
	f, err := os.Open(dst)
	if err != nil {
		t.Fatalf("open generated WebP: %v", err)
	}
	defer func() { _ = f.Close() }()
	_, format, err := image.Decode(f)
	if err != nil {
		t.Fatalf("decode generated WebP: %v", err)
	}
	if format != "webp" {
		t.Fatalf("decoded format = %q, want webp", format)
	}
}

func TestResizeRejectsInvalidInput(t *testing.T) {
	if err := ResizeImage(filepath.Join(t.TempDir(), "missing.png"), filepath.Join(t.TempDir(), "out.png"), 100, 100, "contain", false); err == nil {
		t.Fatal("expected missing source image to error")
	}
}
