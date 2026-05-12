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
		cfg.Log.Debug("Skipping %s: annotation already exists", path)
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
		cfg.Log.Info("[DRY-RUN] Would annotate: %s", path)
		return nil
	}

	if err := atomicWrite(path, []byte(newContent)); err != nil {
		return fmt.Errorf("atomic write error: %w", err)
	}

	cfg.Log.Success("%s", path)

	return nil
}

func atomicWrite(path string, data []byte) error {
	tmpPath := path + ".tmp." + fmt.Sprint(os.Getpid())

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
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
