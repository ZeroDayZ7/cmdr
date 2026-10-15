package profiles

import (
	"os"
	"path/filepath"
	"slices"
)

func DetectProjectProfile(root string, cfg *Config) *Profile {
	// 1. Znajdź realny root projektu idąc w górę drzewa katalogów
	projectRoot := findProjectRootForDetection(root, cfg)

	var bestMatch *Profile
	maxScore := -1

	for _, p := range cfg.Profiles {
		score := 0

		for _, f := range p.Detect.Files {
			if _, err := os.Stat(filepath.Join(projectRoot, f)); err == nil {
				score += 50
			}
		}

		for _, d := range p.Detect.Folders {
			if _, err := os.Stat(filepath.Join(projectRoot, d)); err == nil {
				score += 20
			}
		}

		if score == 0 && len(p.Detect.Extensions) > 0 {
			files, _ := os.ReadDir(projectRoot)
			for _, f := range files {
				if slices.Contains(p.Detect.Extensions, filepath.Ext(f.Name())) {
					score += 10
				}
				if score > 0 {
					break
				}
			}
		}

		if score > maxScore && score > 0 {
			maxScore = score
			pCopy := p
			bestMatch = &pCopy
		}
	}

	return bestMatch
}

// Pomocnicza funkcja szukająca root'a projektu w górę
func findProjectRootForDetection(startPath string, cfg *Config) string {
	current, err := filepath.Abs(startPath)
	if err != nil {
		return startPath
	}
	info, err := os.Stat(current)
	if err == nil && !info.IsDir() {
		current = filepath.Dir(current)
	}

	for {
		// Sprawdź czy w obecnym katalogu znajduje się którykolwiek plik definicji profilu
		for _, p := range cfg.Profiles {
			for _, f := range p.Detect.Files {
				if _, err := os.Stat(filepath.Join(current, f)); err == nil {
					return current
				}
			}
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return startPath
}
