package annotate_region

import (
	"bufio"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/zerodayz7/cmdr/internal/profiles"
)

func ShouldIgnore(path string, info os.FileInfo, customIgnore []string, global profiles.GlobalConfig) bool {
	name := info.Name()

	if info.IsDir() {
		if slices.Contains(global.Ignore, name) {
			return true
		}
	}

	if !info.IsDir() {
		ext := filepath.Ext(path)
		if slices.Contains(global.IgnoredExtensions, ext) {
			return true
		}
	}

	if slices.Contains(customIgnore, name) {
		return true
	}

	return false
}

func LoadIgnoreFile() []string {
	file, err := os.Open(".cmdr_reg_ignore")
	if err != nil {
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil
	}

	return lines
}
