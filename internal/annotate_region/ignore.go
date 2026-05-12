package annotate

import (
	"bufio"
	"os"
	"slices"
	"strings"
)

func ShouldIgnore(path string, info os.FileInfo, customIgnore []string) bool {
	if info.IsDir() {
		if DefaultIgnoredDirs[info.Name()] {
			return true
		}
		if slices.Contains(customIgnore, info.Name()) {
			return true
		}
	}
	return false
}

func LoadIgnoreFile() []string {
	file, err := os.Open(".cmdr_reg_ignore") // Używamy Twojej nazwy pliku ignore
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
	return lines
}
