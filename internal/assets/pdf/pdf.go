package pdf

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	fpdf "github.com/go-pdf/fpdf"
	imageutil "github.com/zerodayz7/cmdr/internal/assets/image"
)

type Options struct {
	OutputName   string
	PageSize     string
	Orientation  string
	MarginMM     float64
	Recursive    bool
	Overwrite    bool
	Preset       string
	MaxSizeBytes int64
	qualityLevel QualityLevel
}

type ImageFile struct {
	Path   string
	Name   string
	Number int64
	Format string
}

type Result struct {
	OutputPath          string
	Files               []ImageFile
	Formats             map[string]int
	Preset              string
	OriginalImageBytes  int64
	OptimizedImageBytes int64
	PDFBytes            int64
	SizeLimitReached    bool
	QualityDPI          int
	JPEGQuality         int
	BackupPath          string
}

func isSupportedFormat(path string) (string, bool) {
	base := filepath.Ext(path)
	switch strings.ToLower(base) {
	case ".png":
		return "PNG", true
	case ".jpg", ".jpeg":
		return "JPG", true
	case ".webp":
		return "WebP", true
	default:
		return "", false
	}
}

func leadingNumber(name string) (int64, bool) {
	stem := filepath.Base(name)
	var digits strings.Builder
	for _, r := range stem {
		if r < '0' || r > '9' {
			break
		}
		digits.WriteRune(r)
	}
	if digits.Len() == 0 {
		return 0, false
	}
	v, err := strconv.ParseInt(digits.String(), 10, 64)
	return v, err == nil
}

func CollectImages(directory string, recursive bool) ([]ImageFile, []string, error) {
	if strings.TrimSpace(directory) == "" {
		return nil, nil, errors.New("source directory is required")
	}
	info, err := os.Stat(directory)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, fmt.Errorf("source directory does not exist: %s", directory)
		}
		return nil, nil, fmt.Errorf("access source directory %s: %w", directory, err)
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("source path is not a directory: %s", directory)
	}

	files := make([]ImageFile, 0)
	unnumbered := make([]string, 0)
	walk := func(path string, entry os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if !recursive && path != filepath.Clean(directory) {
				return filepath.SkipDir
			}
			return nil
		}
		format, supported := isSupportedFormat(path)
		if !supported {
			return nil
		}
		n, numbered := leadingNumber(filepath.Base(path))
		if !numbered {
			unnumbered = append(unnumbered, filepath.Base(path))
			return nil
		}
		files = append(files, ImageFile{Path: path, Name: filepath.Base(path), Number: n, Format: format})
		return nil
	}
	if err := filepath.Walk(directory, walk); err != nil {
		return nil, nil, fmt.Errorf("scan source directory %s: %w", directory, err)
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].Number != files[j].Number {
			return files[i].Number < files[j].Number
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})
	for i := 1; i < len(files); i++ {
		if files[i].Number == files[i-1].Number {
			return nil, unnumbered, fmt.Errorf("duplicate numeric prefix %d: %s and %s", files[i].Number, files[i-1].Name, files[i].Name)
		}
	}
	if len(files) == 0 {
		if len(unnumbered) > 0 {
			return nil, unnumbered, errors.New("no numbered supported images found")
		}
		return nil, unnumbered, errors.New("no supported images found in source directory")
	}
	return files, unnumbered, nil
}

func Generate(directory string, options Options) (Result, error) {
	preset, err := NormalizePreset(options.Preset)
	if err != nil {
		return Result{}, err
	}
	profile, ok := GetProfile(preset)
	if !ok {
		return Result{}, fmt.Errorf("invalid preset %q", preset)
	}
	base := filepath.Clean(directory)
	outputName := strings.TrimSpace(options.OutputName)
	if outputName == "" {
		name := filepath.Base(base)
		if name == "." || name == string(filepath.Separator) {
			name = "cmdr"
		}
		outputName = name + ".pdf"
	}
	if !strings.HasSuffix(strings.ToLower(outputName), ".pdf") {
		outputName += ".pdf"
	}
	var outputPath string
	if filepath.IsAbs(outputName) || filepath.Dir(outputName) != "." {
		outputPath = filepath.Clean(outputName)
	} else {
		outputPath = filepath.Join(base, outputName)
	}
	if _, err := os.Stat(outputPath); err == nil && !options.Overwrite {
		return Result{}, fmt.Errorf("output file already exists: %s (use --force to overwrite)", outputPath)
	} else if err != nil && !os.IsNotExist(err) {
		return Result{}, fmt.Errorf("check output file %s: %w", outputPath, err)
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return Result{}, fmt.Errorf("create output directory: %w", err)
	}

	var finalResult Result
	for index, level := range profile.Levels {
		candidate, err := os.CreateTemp(filepath.Dir(outputPath), ".cmdr-candidate-*.pdf")
		if err != nil {
			return Result{}, fmt.Errorf("create candidate PDF: %w", err)
		}
		candidatePath := candidate.Name()
		_ = candidate.Close()
		_ = os.Remove(candidatePath)

		attempt := options
		attempt.Preset = preset
		attempt.OutputName = candidatePath
		attempt.Overwrite = true
		attempt.qualityLevel = level
		result, err := generateOnce(directory, attempt)
		if err != nil {
			_ = os.Remove(candidatePath)
			return Result{}, err
		}
		info, err := os.Stat(candidatePath)
		if err != nil {
			_ = os.Remove(candidatePath)
			return Result{}, fmt.Errorf("measure generated PDF: %w", err)
		}
		if !info.Mode().IsRegular() {
			_ = os.Remove(candidatePath)
			return Result{}, fmt.Errorf("generated PDF is not a regular file: %s", candidatePath)
		}
		result.PDFBytes = info.Size()
		result.Preset = preset
		result.QualityDPI = level.DPI
		result.JPEGQuality = level.JPEGQuality
		result.SizeLimitReached = options.MaxSizeBytes <= 0 || result.PDFBytes <= options.MaxSizeBytes
		finalResult = result

		if options.MaxSizeBytes <= 0 || result.PDFBytes <= options.MaxSizeBytes {
			break
		}
		if index < len(profile.Levels)-1 {
			_ = os.Remove(candidatePath)
		}
	}

	candidatePath := finalResult.OutputPath
	if options.Overwrite {
		if _, err := os.Stat(outputPath); err == nil {
			backup, err := os.CreateTemp(filepath.Dir(outputPath), ".cmdr-backup-*.pdf")
			if err != nil {
				_ = os.Remove(candidatePath)
				return Result{}, fmt.Errorf("create PDF backup: %w", err)
			}
			backupPath := backup.Name()
			_ = backup.Close()
			_ = os.Remove(backupPath)
			if err := os.Rename(outputPath, backupPath); err != nil {
				_ = os.Remove(candidatePath)
				return Result{}, fmt.Errorf("prepare existing PDF backup: %w", err)
			}
			if err := os.Rename(candidatePath, outputPath); err != nil {
				rollbackErr := os.Rename(backupPath, outputPath)
				_ = os.Remove(candidatePath)
				if rollbackErr != nil {
					return Result{}, fmt.Errorf("save output PDF %s: %w; rollback failed; previous PDF preserved as %s: %v", outputPath, err, backupPath, rollbackErr)
				}
				return Result{}, fmt.Errorf("save output PDF %s: %w; existing PDF restored", outputPath, err)
			}
			_ = os.Remove(backupPath)
			finalResult.OutputPath = outputPath
			return finalResult, nil
		} else if !os.IsNotExist(err) {
			_ = os.Remove(candidatePath)
			return Result{}, fmt.Errorf("check existing output PDF: %w", err)
		}
	}
	if err := os.Rename(candidatePath, outputPath); err != nil {
		_ = os.Remove(candidatePath)
		return Result{}, fmt.Errorf("save output PDF %s: %w", outputPath, err)
	}
	finalResult.OutputPath = outputPath
	return finalResult, nil
}

func generateOnce(directory string, options Options) (Result, error) {
	files, _, err := CollectImages(directory, options.Recursive)
	if err != nil {
		return Result{}, err
	}

	base := filepath.Clean(directory)
	if info, err := os.Stat(base); err != nil {
		return Result{}, fmt.Errorf("access source directory: %w", err)
	} else if !info.IsDir() {
		return Result{}, fmt.Errorf("source path is not a directory: %s", directory)
	}

	outputName := strings.TrimSpace(options.OutputName)
	if outputName == "" {
		name := filepath.Base(filepath.Clean(directory))
		if name == "." || name == string(filepath.Separator) {
			name = "cmdr"
		}
		outputName = name + ".pdf"
	}
	if !strings.HasSuffix(strings.ToLower(outputName), ".pdf") {
		outputName += ".pdf"
	}
	var outputPath string
	if filepath.IsAbs(outputName) || filepath.Dir(outputName) != "." {
		outputPath = filepath.Clean(outputName)
	} else {
		outputPath = filepath.Join(base, outputName)
	}

	if _, err := os.Stat(outputPath); err == nil && !options.Overwrite {
		return Result{}, fmt.Errorf("output file already exists: %s (use --force to overwrite)", outputPath)
	} else if err != nil && !os.IsNotExist(err) {
		return Result{}, fmt.Errorf("check output file %s: %w", outputPath, err)
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return Result{}, fmt.Errorf("create output directory: %w", err)
	}

	pageSize := strings.ToUpper(strings.TrimSpace(options.PageSize))
	if pageSize == "" {
		pageSize = "A4"
	}
	var size fpdf.SizeType
	switch pageSize {
	case "A4":
		size = fpdf.SizeType{Wd: 210, Ht: 297}
	case "LETTER":
		size = fpdf.SizeType{Wd: 215.9, Ht: 279.4}
	default:
		return Result{}, fmt.Errorf("unsupported page size %q: expected A4 or LETTER", pageSize)
	}
	orientation := strings.ToLower(strings.TrimSpace(options.Orientation))
	if orientation == "" {
		orientation = "auto"
	}
	if orientation != "auto" && orientation != "portrait" && orientation != "landscape" {
		return Result{}, fmt.Errorf("invalid orientation %q: expected auto, portrait, or landscape", orientation)
	}
	if options.MarginMM < 0 {
		return Result{}, fmt.Errorf("margin must be zero or positive millimeters")
	}

	tmp, err := os.CreateTemp(filepath.Dir(outputPath), ".cmdr-pdf-*.tmp")
	if err != nil {
		return Result{}, fmt.Errorf("create temporary PDF: %w", err)
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()
	cleanup := func() { _ = os.Remove(tmpPath) }

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(true)
	formats := make(map[string]int)
	originalImageBytes, err := originalImageBytes(files)
	if err != nil {
		cleanup()
		return Result{}, fmt.Errorf("measure source images: %w", err)
	}
	optimizedImageBytes := int64(0)
	for _, file := range files {
		img, _, err := imageutil.DecodeImage(file.Path)
		if err != nil {
			cleanup()
			return Result{}, fmt.Errorf("decode image %s: %w", file.Name, err)
		}
		bounds := img.Bounds()
		if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
			cleanup()
			return Result{}, fmt.Errorf("decoded image %s has invalid dimensions", file.Name)
		}

		level := options.qualityLevel
		if level.DPI == 0 && level.JPEGQuality == 0 {
			profile, ok := GetProfile(normalizePresetForFallback(options.Preset))
			if ok {
				level = profile.Levels[0]
			}
		}
		pageForImage := fpdfSize{w: size.Wd, h: size.Ht}
		actualOrientation := orientation
		if orientation == "auto" {
			if bounds.Dx() > bounds.Dy() {
				actualOrientation = "landscape"
			} else {
				actualOrientation = "portrait"
			}
		}
		if actualOrientation == "landscape" {
			pageForImage.w, pageForImage.h = size.Ht, size.Wd
		} else {
			pageForImage.w, pageForImage.h = size.Wd, size.Ht
		}
		pageForImage.w -= 2 * options.MarginMM
		pageForImage.h -= 2 * options.MarginMM
		if pageForImage.w < 1 {
			pageForImage.w = 1
		}
		if pageForImage.h < 1 {
			pageForImage.h = 1
		}
		optimized, err := optimizeImageWithSize(img, normalizePresetForFallback(options.Preset), level, pageForImage)
		if err != nil {
			cleanup()
			return Result{}, fmt.Errorf("optimize image %s: %w", file.Name, err)
		}
		optimizedImageBytes += int64(len(optimized.Data))

		pdf.AddPageFormat(actualOrientation, size)
		pageW, pageH := pdf.GetPageSize()
		contentW := pageW - 2*options.MarginMM
		contentH := pageH - 2*options.MarginMM
		if contentW <= 0 || contentH <= 0 {
			cleanup()
			return Result{}, errors.New("margin leaves no usable page area")
		}
		w, h := float64(optimized.Width), float64(optimized.Height)
		scale := contentW / w
		if contentH/h < scale {
			scale = contentH / h
		}
		drawW, drawH := w*scale, h*scale
		x := options.MarginMM + (contentW-drawW)/2
		y := options.MarginMM + (contentH-drawH)/2
		pdf.RegisterImageOptionsReader(file.Name, fpdf.ImageOptions{ImageType: optimized.Format, ReadDpi: false}, bytes.NewReader(optimized.Data))
		pdf.ImageOptions(file.Name, x, y, drawW, drawH, false, fpdf.ImageOptions{ImageType: optimized.Format}, 0, "")
		if pdf.Error() != nil {
			cleanup()
			return Result{}, fmt.Errorf("generate PDF page for %s: %w", file.Name, pdf.Error())
		}
		formats[file.Format]++
	}

	if err := pdf.OutputFileAndClose(tmpPath); err != nil {
		cleanup()
		return Result{}, fmt.Errorf("generate PDF: %w", err)
	}
	if err := os.Rename(tmpPath, outputPath); err != nil {
		if options.Overwrite {
			if removeErr := os.Remove(outputPath); removeErr != nil && !os.IsNotExist(removeErr) {
				cleanup()
				return Result{}, fmt.Errorf("replace output PDF %s: %w", outputPath, removeErr)
			}
			if renameErr := os.Rename(tmpPath, outputPath); renameErr != nil {
				cleanup()
				return Result{}, fmt.Errorf("save output PDF %s: %w", outputPath, renameErr)
			}
		} else {
			cleanup()
			return Result{}, fmt.Errorf("save output PDF %s: %w", outputPath, err)
		}
	}
	return Result{OutputPath: outputPath, Files: files, Formats: formats, OriginalImageBytes: originalImageBytes, OptimizedImageBytes: optimizedImageBytes}, nil
}
