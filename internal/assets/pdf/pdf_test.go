package pdf

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/chai2010/webp"
)

func writeImage(t *testing.T, path string, width, height int, format string) {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 40, G: uint8(x), B: uint8(y), A: 255})
		}
	}
	if format == "png" {
		for y := height / 3; y < 2*height/3; y++ {
			for x := width / 3; x < 2*width/3; x++ {
				img.SetNRGBA(x, y, color.NRGBA{R: 255, G: 0, B: 0, A: 80})
			}
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	closeFile := func() {
		if closeErr := f.Close(); closeErr != nil {
			t.Fatalf("close test image %s: %v", path, closeErr)
		}
	}
	switch format {
	case "png":
		err = png.Encode(f, img)
	case "jpg":
		err = jpeg.Encode(f, img, &jpeg.Options{Quality: 92})
	case "webp":
		err = webp.Encode(f, img, &webp.Options{Quality: 90, Lossless: true, Exact: true})
	default:
		_ = f.Close()
		t.Fatalf("unsupported test format %s", format)
	}
	if err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	closeFile()
}

func TestCollectImagesNumericSortMixedExtensions(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"10_final.jpg", "002_middle.webp", "1_cover.png", "3_page.png", "11_final2.PNG"} {
		writeImage(t, filepath.Join(dir, name), 20, 30, "png")
	}
	files, unnumbered, err := CollectImages(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(unnumbered) != 0 {
		t.Fatalf("unnumbered = %v", unnumbered)
	}
	want := []int64{1, 2, 3, 10, 11}
	for i, file := range files {
		if file.Number != want[i] {
			t.Fatalf("sort[%d] = %d, want %d", i, file.Number, want[i])
		}
	}
}

func TestCollectImagesLeadingZeroSort(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"010_final.jpg", "001_okladka.png", "002_intro.webp"} {
		writeImage(t, filepath.Join(dir, name), 20, 30, "png")
	}
	files, _, err := CollectImages(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	want := []int64{1, 2, 10}
	for i, file := range files {
		if file.Number != want[i] {
			t.Fatalf("sort[%d] = %d, want %d", i, file.Number, want[i])
		}
	}
}

func TestDuplicateAndUnnumbered(t *testing.T) {
	dir := t.TempDir()
	writeImage(t, filepath.Join(dir, "001.png"), 10, 10, "png")
	writeImage(t, filepath.Join(dir, "001_copy.png"), 10, 10, "png")
	if _, _, err := CollectImages(dir, false); err == nil {
		t.Fatal("expected duplicate numeric prefix error")
	}

	dir = t.TempDir()
	writeImage(t, filepath.Join(dir, "cover.png"), 10, 10, "png")
	files, skipped, err := CollectImages(dir, false)
	if err == nil {
		t.Fatal("expected no numbered images error")
	}
	if len(files) != 0 || len(skipped) != 1 {
		t.Fatalf("files=%d skipped=%v", len(files), skipped)
	}
}

func TestGenerateMixedOrientationsAndTransparency(t *testing.T) {
	dir := t.TempDir()
	writeImage(t, filepath.Join(dir, "001_cover.png"), 40, 60, "png")
	writeImage(t, filepath.Join(dir, "002_spread.jpg"), 80, 40, "jpg")
	writeImage(t, filepath.Join(dir, "010_last.webp"), 30, 30, "webp")
	result, err := Generate(dir, Options{PageSize: "A4", Orientation: "auto", MarginMM: 0, Overwrite: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Files) != 3 {
		t.Fatalf("files=%d", len(result.Files))
	}
	data, err := os.ReadFile(result.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	pages := regexp.MustCompile(`/Type\s*/Page\b`).FindAll(data, -1)
	if len(pages) != 3 {
		t.Fatalf("page objects=%d, want 3", len(pages))
	}
}

func TestGenerateCorruptExistsAndForce(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "001_bad.png"), []byte("not an image"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(dir, Options{}); err == nil {
		t.Fatal("expected corrupt image error")
	}

	clean := t.TempDir()
	writeImage(t, filepath.Join(clean, "1.png"), 20, 20, "png")
	first, err := Generate(clean, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(clean, Options{}); err == nil {
		t.Fatal("expected existing output error")
	}
	if _, err := Generate(clean, Options{Overwrite: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(first.OutputPath); err != nil {
		t.Fatal(err)
	}
}

func TestEmptyAndInvalidDirectory(t *testing.T) {
	if _, _, err := CollectImages(filepath.Join(t.TempDir(), "missing"), false); err == nil {
		t.Fatal("expected missing directory error")
	}
	if _, _, err := CollectImages(t.TempDir(), false); err == nil {
		t.Fatal("expected empty directory error")
	}
}
