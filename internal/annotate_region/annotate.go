package annotate

import (
	"os"
	"path/filepath"
	"strings"
)

func Process(targetPath string, cfg Config) error {
	info, err := os.Stat(targetPath)
	if err != nil {
		return err
	}

	customIgnore := LoadIgnoreFile()

	if !info.IsDir() {
		if strings.HasSuffix(targetPath, ".go") {
			// POPRAWKA: Przekazujemy cfg
			_, err = AnnotateRegions(targetPath, cfg)
			return err
		}
		return nil
	}

	return filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if ShouldIgnore(path, info, customIgnore) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			// POPRAWKA: Tutaj już miałeś cfg, teraz sygnatura w regions.go się zgadza
			_, err = AnnotateRegions(path, cfg)
			return err
		}
		return nil
	})
}
