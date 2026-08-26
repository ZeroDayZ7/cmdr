package annotate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/zerodayz7/cmdr/internal/profiles"
)

var shebangRegex = regexp.MustCompile(`^#!\s*/.+`)

type Logger interface {
	Success(msg string, args ...any)
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

type Config struct {
	DryRun         bool
	Verbose        bool
	Profile        *profiles.Profile
	ProfilesConfig *profiles.Config
	Log            Logger
}

func AnnotateFile(path string, cfg Config, commentGen func(relPath string, style string) string) error {
	ext := strings.ToLower(filepath.Ext(path))

	style, ok := cfg.ProfilesConfig.CommentStyles[ext]
	if !ok {
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
		return nil
	}

	root, _ := FindProjectRoot(path, cfg)
	relPath, err := filepath.Rel(root, path)
	if err != nil {
		relPath = filepath.Base(path)
	}

	comment := commentGen(relPath, style)
	newContent := InjectComment(content, comment)

	if cfg.DryRun {
		cfg.Log.Info("[DRY-RUN] Would annotate: %s", relPath)
		return nil
	}

	return AtomicWrite(path, []byte(newContent))
}

func AtomicWrite(path string, data []byte) error {
	tmpPath := path + ".tmp." + fmt.Sprint(os.Getpid())
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}
	return nil
}

func InjectComment(content, comment string) string {
	if len(strings.TrimSpace(content)) == 0 {
		return comment
	}
	lines := strings.SplitN(content, "\n", 2)
	if len(lines) > 0 && shebangRegex.MatchString(lines[0]) {
		if len(lines) > 1 {
			return lines[0] + "\n" + comment + "\n" + lines[1]
		}
		return lines[0] + "\n" + comment
	}
	return comment + "\n\n" + content
}

func ShouldIgnore(path string, cfg Config) bool {
	if cfg.ProfilesConfig == nil {
		return false
	}

	cleanPath := filepath.ToSlash(path)
	ignores := cfg.ProfilesConfig.Global.Ignore
	if cfg.Profile != nil {
		ignores = append(ignores, cfg.Profile.Ignore...)
	}

	for _, p := range ignores {
		p = filepath.ToSlash(p)
		if strings.Contains(cleanPath, "/"+p+"/") ||
			strings.HasSuffix(cleanPath, "/"+p) ||
			strings.HasPrefix(cleanPath, p+"/") {
			return true
		}
	}
	return false
}

func IsAllowedExtension(ext string, cfg Config) bool {
	if cfg.Profile != nil && len(cfg.Profile.Extensions) > 0 {
		for _, e := range cfg.Profile.Extensions {
			if strings.EqualFold(e, ext) {
				return true
			}
		}
		return false
	}

	if cfg.ProfilesConfig == nil || cfg.ProfilesConfig.CommentStyles == nil {
		return strings.EqualFold(ext, ".go")
	}

	_, ok := cfg.ProfilesConfig.CommentStyles[strings.ToLower(ext)]
	return ok
}

func IsBinary(data []byte) bool {
	limit := min(len(data), 1024)

	for i := range limit {
		if data[i] == 0 {
			return true
		}
	}

	return !utf8.Valid(data[:limit])
}

//#region HasAnnotation
func HasAnnotation(content string) bool {
	lines := strings.SplitN(content, "\n", 3)
	if len(lines) == 0 {
		return false
	}

	firstLine := strings.TrimSpace(lines[0])

	// Sprawdzamy czy pierwsza linia jest komentarzem jednolinkowym (//, #, <!--, /*)
	return strings.HasPrefix(firstLine, "//") ||
		strings.HasPrefix(firstLine, "#") ||
		strings.HasPrefix(firstLine, "<!--") ||
		strings.HasPrefix(firstLine, "/*")
}
//#endregion

func FindProjectRoot(path string, cfg Config) (string, error) {
	current := filepath.Dir(path)
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}
		if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return filepath.Dir(path), nil
}
