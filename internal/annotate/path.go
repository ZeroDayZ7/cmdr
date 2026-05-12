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

func FindProjectRoot(path string) (string, error) {
	curr, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if !isDir(curr) {
		curr = filepath.Dir(curr)
	}

	for {
		for _, marker := range ProjectMarkers {
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
	rel, err := filepath.Rel(root, file)
	if err != nil {
		return "", err
	}

	project := filepath.Base(root)

	return filepath.ToSlash(
		filepath.Join(project, rel),
	), nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}
