package annotate_region

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zerodayz7/cmdr/internal/profiles"
)

func AnnotateRegions(filePath string, cfg Config, profile *profiles.Profile, fullCfg *profiles.Config) (bool, error) {
	if profile == nil || len(profile.RegionPatterns) == 0 {
		return false, nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	ext := filepath.Ext(filePath)
	commentStyle, ok := fullCfg.CommentStyles[ext]
	if !ok {
		commentStyle = "// %s"
	}

	strContent := string(content)
	currentContent := strContent
	modified := false

	for _, p := range profile.RegionPatterns {
		re, err := regexp.Compile("(?m)" + p.Regex)
		if err != nil {
			continue
		}

		currentContent = re.ReplaceAllStringFunc(currentContent, func(match string) string {
			submatch := re.FindStringSubmatch(match)
			if len(submatch) < 2 {
				return match
			}

			name := submatch[1]
			regionLabel := fmt.Sprintf("#region %s", name)
			regionTag := fmt.Sprintf(commentStyle, regionLabel)

			if strings.Contains(strContent, regionTag) {
				return match
			}

			modified = true
			return fmt.Sprintf("%s\n%s", regionTag, match)
		})
	}

	if modified {
		if cfg.Verbose {
			fmt.Printf("[ANNOTATE][%s] Found new regions in: %s\n", profile.Name, filePath)
		}

		if cfg.DryRun {
			return true, nil
		}

		err = os.WriteFile(filePath, []byte(currentContent), 0644)
	}

	return modified, err
}
