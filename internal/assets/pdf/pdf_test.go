package pdf

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

func TestProfilesAndParseSize(t *testing.T) {
	for _, preset := range []string{"print", "balanced", "ai", "preview"} {
		if _, err := NormalizePreset(preset); err != nil {
			t.Fatalf("preset %s: %v", preset, err)
		}
	}
	if _, err := NormalizePreset("invalid"); err == nil {
		t.Fatal("expected invalid preset error")
	}
	cases := map[string]int64{"800KB": 800000, "2MB": 2000000, "64KIB": 65536, "100B": 100}
	for input, want := range cases {
		got, err := ParseSize(input)
		if err != nil || got != want {
			t.Fatalf("ParseSize(%s) = %d, %v; want %d", input, got, err, want)
		}
	}
}

func TestParseSizeRejectsInvalidAndOverflow(t *testing.T) {
	for _, input := range []string{"0KB", "-1MB", "1XB", "999999999999999999999999MB"} {
		if _, err := ParseSize(input); err == nil {
			t.Fatalf("ParseSize(%q) should fail", input)
		}
	}
	if got, err := ParseSize(" 800KB "); err != nil || got != 800000 {
		t.Fatalf("ParseSize whitespace = %d, %v", got, err)
	}
}

func TestResizePreservesAspectRatioAndAlphaComposites(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 400, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 400; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 200, G: 30, B: 80, A: uint8(x % 256)})
		}
	}
	resized, w, h := resizeToDPI(img, 96, fpdfSize{w: 210, h: 297})
	if w != 400 || h != 200 {
		if float64(w)/float64(h) < 1.9 || float64(w)/float64(h) > 2.1 {
			t.Fatalf("aspect ratio changed: %dx%d", w, h)
		}
	}
	level := QualityLevel{DPI: 96, JPEGQuality: 70}
	data, typ, err := optimizeImage(img, "ai", level, fpdfSize{w: 210, h: 297})
	if err != nil || typ != "jpg" || len(data) == 0 {
		t.Fatalf("optimized image: type=%s size=%d err=%v", typ, len(data), err)
	}
	if resized.Bounds().Empty() {
		t.Fatal("resized image is empty")
	}
}

func TestPrintPresetKeeepsPNGAndPreservesSource(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "001_source.png")
	writeImage(t, path, 40, 60, "png")
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	result, err := Generate(dir, Options{Preset: "print", PageSize: "A4", Orientation: "auto", Overwrite: false})
	if err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("print preset modified source image")
	}
	if result.Preset != "print" || result.PDFBytes <= 0 {
		t.Fatalf("unexpected print result: %+v", result)
	}
}

func TestMaxSizeUnreachableReportsMinimumLevel(t *testing.T) {
	dir := t.TempDir()
	writeImage(t, filepath.Join(dir, "001_image.png"), 800, 800, "png")
	result, err := Generate(dir, Options{Preset: "preview", MaxSizeBytes: 1, Overwrite: false})
	if err != nil {
		t.Fatal(err)
	}
	if result.SizeLimitReached {
		t.Fatal("tiny 1-byte limit unexpectedly reached")
	}
	if result.QualityDPI != 72 {
		t.Fatalf("minimum preview DPI = %d, want 72", result.QualityDPI)
	}
}

func TestDPIUsesLETTERAndMargins(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1200, 800))
	letterPortrait := fpdfSize{w: 215.9 - 10, h: 279.4 - 10}
	resized, w, h := resizeToDPI(img, 96, letterPortrait)
	if resized == nil || w <= 0 || h <= 0 {
		t.Fatal("LETTER DPI resize returned empty dimensions")
	}
	if float64(w)/float64(h) < 1.49 || float64(w)/float64(h) > 1.51 {
		t.Fatalf("LETTER resize changed aspect ratio: %dx%d", w, h)
	}
}

func TestAlphaCompositesToWhiteWithoutBoundsShift(t *testing.T) {
	img := image.NewNRGBA(image.Rect(10, 20, 12, 22))
	img.SetNRGBA(10, 20, color.NRGBA{R: 255, G: 0, B: 0, A: 128})
	img.SetNRGBA(11, 21, color.NRGBA{R: 0, G: 0, B: 255, A: 0})
	optimized, err := optimizeImageWithSize(img, "ai", QualityLevel{DPI: 96, JPEGQuality: 100}, fpdfSize{w: 210, h: 297})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := jpeg.Decode(bytes.NewReader(optimized.Data))
	if err != nil {
		t.Fatal(err)
	}
	red := color.RGBAModel.Convert(decoded.At(0, 0)).(color.RGBA)
	if red.R <= red.G || red.R <= red.B {
		t.Fatalf("unexpected composited red pixel: %+v", red)
	}
	corner := color.RGBAModel.Convert(decoded.At(1, 1)).(color.RGBA)
	if corner.R < 245 || corner.G < 245 || corner.B < 245 {
		t.Fatalf("transparent pixel was not white: %+v", corner)
	}
}

func TestForceReplacementLeavesNoTemporaryPDFs(t *testing.T) {
	dir := t.TempDir()
	writeImage(t, filepath.Join(dir, "001_image.png"), 40, 40, "png")
	if _, err := Generate(dir, Options{Preset: "print", Overwrite: false}); err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(dir, Options{Preset: "ai", Overwrite: true}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.Contains(name, ".cmdr-candidate-") || strings.Contains(name, ".cmdr-backup-") {
			t.Fatalf("temporary PDF remains: %s", name)
		}
	}
}
