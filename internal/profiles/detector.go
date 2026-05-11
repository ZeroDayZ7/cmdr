package profiles

import (
	"os"
	"path/filepath"
)

func DetectProjectProfile(root string, cfg *Config) *Profile {
	var bestMatch *Profile
	maxScore := -1

	for _, p := range cfg.Profiles {
		score := 0

		
		for _, f := range p.Detect.Files {
			if _, err := os.Stat(filepath.Join(root, f)); err == nil {
				score += 50
			}
		}

		
		for _, d := range p.Detect.Folders {
			if _, err := os.Stat(filepath.Join(root, d)); err == nil {
				score += 20
			}
		}

		
		if score == 0 && len(p.Detect.Extensions) > 0 {
			files, _ := os.ReadDir(root)
			for _, f := range files {
				for _, ext := range p.Detect.Extensions {
					if filepath.Ext(f.Name()) == ext {
						score += 10
						break
					}
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
