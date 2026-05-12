package annotate

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Obsługuje funkcje zwykłe oraz metody na strukturach
var funcRegex = regexp.MustCompile(`(?m)^func\s+(?:\([^\)]+\)\s+)?([a-zA-Z0-9_]+)\s*\(`)

// Zaktualizowana sygnatura: przyjmuje Config
func AnnotateRegions(filePath string, cfg Config) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	strContent := string(content)
	modified := false

	newContent := funcRegex.ReplaceAllStringFunc(strContent, func(match string) string {
		submatch := funcRegex.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}

		funcName := submatch[1]
		regionTag := fmt.Sprintf("// #region %s", funcName)

		if strings.Contains(strContent, regionTag) {
			return match
		}

		modified = true
		return fmt.Sprintf("%s\n%s", regionTag, match)
	})

	if modified {
		// Logowanie verbose
		if cfg.Verbose {
			fmt.Printf("[ANNOTATE] Found new regions in: %s\n", filePath)
		}

		// Respektowanie DryRun
		if cfg.DryRun {
			return true, nil
		}

		err = os.WriteFile(filePath, []byte(newContent), 0644)
	}

	return modified, err
}
