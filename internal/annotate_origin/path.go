package annotate

import (
	"os"
	"path/filepath"
)

var ProjectMarkers = []string{
	".git",
	"go.mod",
	"package.json",
	"pubspec.yaml",
}

func FindProjectRoot(path string, cfg Config) (string, error) {
	curr, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if !isDir(curr) {
		curr = filepath.Dir(curr)
	}

	markers := ProjectMarkers

	if cfg.Profile != nil && len(cfg.Profile.Detect.Files) > 0 {
		markers = append(cfg.Profile.Detect.Files, markers...)
	}

	for {
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(curr, marker)); err == nil {
				return curr, nil
			}
		}

		parent := filepath.Dir(curr)
		if parent == curr {
			return "", os.ErrNotExist
		}
		curr = parent
	}
}

func BuildRelativePath(root, file string) (string, error) {

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	absFile, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(absRoot, absFile)
	if err != nil {
		return "", err
	}

	project := filepath.Base(absRoot)

	return filepath.ToSlash(filepath.Join(project, rel)), nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {

		return false
	}
	return info.IsDir()
}
