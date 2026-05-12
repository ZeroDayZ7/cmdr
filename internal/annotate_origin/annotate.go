package annotate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func AnnotateFile(path string, cfg Config) error {
	ext := strings.ToLower(filepath.Ext(path))

	if !isAllowedExtension(ext, cfg) {
		return nil
	}

	style, ok := cfg.ProfilesConfig.CommentStyles[ext]
	if !ok {
		return nil
	}

	if shouldIgnore(path, cfg) {
		return nil
	}

	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read error: %w", err)
	}

	if IsBinary(contentBytes) {
		return nil
	}

	content := string(contentBytes)

	if HasAnnotation(content) {
		if cfg.Verbose {
			fmt.Printf("[SKIP] %s\n", path)
		}
		return nil
	}

	root, err := FindProjectRoot(path, cfg)
	if err != nil {

		root = filepath.Dir(path)
	}

	relPath, err := BuildRelativePath(root, path)
	if err != nil {
		relPath = filepath.Base(path)
	}

	comment := fmt.Sprintf(style, relPath)

	newContent := injectComment(content, comment)

	if cfg.DryRun {
		fmt.Printf("[ADD] %s\n", path)
		return nil
	}

	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("write error: %w", err)
	}

	if cfg.Verbose {
		fmt.Printf("[OK] %s\n", path)
	}

	return nil
}

var shebangRegex = regexp.MustCompile(`^#!\s*/.+`)

func injectComment(content, comment string) string {
	lines := strings.SplitN(content, "\n", 2)

	if len(lines) > 0 && shebangRegex.MatchString(lines[0]) {
		if len(lines) > 1 {
			return lines[0] + "\n" + comment + "\n" + lines[1]
		}
		return lines[0] + "\n" + comment
	}

	return comment + "\n\n" + content
}
