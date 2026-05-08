package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type AnnotateConfig struct {
	DryRun  bool
	Verbose bool
}

var annotationRegex = regexp.MustCompile(`(?m)^(//|#|--|/\*|<!--)\s*cmdr:`)

var commentStyles = map[string]string{
	".go":    "// cmdr: %s",
	".js":    "// cmdr: %s",
	".jsx":   "// cmdr: %s",
	".ts":    "// cmdr: %s",
	".tsx":   "// cmdr: %s",
	".java":  "// cmdr: %s",
	".kt":    "// cmdr: %s",
	".swift": "// cmdr: %s",
	".c":     "// cmdr: %s",
	".cpp":   "// cmdr: %s",
	".hpp":   "// cmdr: %s",
	".cs":    "// cmdr: %s",
	".php":   "// cmdr: %s",
	".rs":    "// cmdr: %s",

	".yaml": "# cmdr: %s",
	".yml":  "# cmdr: %s",
	".py":   "# cmdr: %s",
	".sh":   "# cmdr: %s",
	".rb":   "# cmdr: %s",
	".toml": "# cmdr: %s",
	".env":  "# cmdr: %s",

	".sql": "-- cmdr: %s",

	".css":  "/* cmdr: %s */",
	".scss": "/* cmdr: %s */",

	".html": "<!-- cmdr: %s -->",
	".md":   "<!-- cmdr: %s -->",
}

var ignoredExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".webp": true,
	".svg":  true,
	".ico":  true,

	".exe": true,
	".dll": true,
	".so":  true,
	".bin": true,

	".zip": true,
	".rar": true,
	".7z":  true,
	".tar": true,
	".gz":  true,

	".pdf": true,
	".mp3": true,
	".mp4": true,
	".mov": true,

	".lock": true,
}

var projectMarkers = []string{
	"go.mod",
	"package.json",
	"pubspec.yaml",
	"Cargo.toml",
	"pom.xml",
	"requirements.txt",
	".git",
}

var ignoredDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"dist":         true,
	"build":        true,
	".next":        true,
	".turbo":       true,
	"coverage":     true,
	"vendor":       true,
	"bin":          true,
	"obj":          true,
	".dart_tool":   true,
	".idea":        true,
	".vscode":      true,
}

func ShouldIgnoreDir(name string) bool {
	if ignoredDirs[name] {
		return true
	}

	customIgnore, _ := ReadIgnoreFile(".annotateignore")
	for _, item := range customIgnore {
		if item == name {
			return true
		}
	}

	return false
}

func findProjectRoot(startPath string) (string, error) {
	curr, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	for {
		for _, marker := range projectMarkers {
			if _, err := os.Stat(filepath.Join(curr, marker)); err == nil {
				return curr, nil
			}
		}

		parent := filepath.Dir(curr)

		if parent == curr {
			return "", fmt.Errorf("project root not found")
		}

		curr = parent
	}
}

func AnnotateFile(path string, cfg AnnotateConfig) error {
	ext := strings.ToLower(filepath.Ext(path))

	if ignoredExtensions[ext] {
		return nil
	}

	style, ok := commentStyles[ext]
	if !ok {
		return nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	root, err := findProjectRoot(absPath)
	if err != nil {
		return nil
	}

	relPath, err := filepath.Rel(root, absPath)
	if err != nil {
		return err
	}

	projectName := filepath.Base(root)
	displayPath := filepath.ToSlash(filepath.Join(projectName, relPath))

	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	content := string(contentBytes)

	scanner := bufio.NewScanner(strings.NewReader(content))

	lineCount := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			lineCount++
			if lineCount > 5 {
				break
			}
			continue
		}

		if annotationRegex.MatchString(line) {
			if cfg.Verbose {
				fmt.Printf("[SKIP] already annotated: %s\n", path)
			}
			return nil
		}

		break
	}

	comment := fmt.Sprintf(style, displayPath)

	var newContent string

	if strings.HasPrefix(content, "#!") {
		parts := strings.SplitN(content, "\n", 2)

		if len(parts) > 1 {
			newContent = parts[0] + "\n" + comment + "\n" + parts[1]
		} else {
			newContent = parts[0] + "\n" + comment
		}
	} else {
		newContent = comment + "\n\n" + content
	}

	if cfg.DryRun {
		fmt.Printf("[ADD] %s\n", path)
		return nil
	}

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return err
	}

	if cfg.Verbose {
		fmt.Printf("[OK] annotated: %s\n", path)
	}

	return nil
}
