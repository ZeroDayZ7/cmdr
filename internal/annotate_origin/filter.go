package annotate

import (
	"path/filepath"
	"strings"
)

func shouldIgnore(path string, cfg Config) bool {
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

func isAllowedExtension(ext string, cfg Config) bool {
	if cfg.Profile != nil && len(cfg.Profile.Extensions) > 0 {
		for _, e := range cfg.Profile.Extensions {
			if strings.EqualFold(e, ext) {
				return true
			}
		}

		return false
	}

	_, ok := cfg.ProfilesConfig.CommentStyles[strings.ToLower(ext)]
	return ok
}
