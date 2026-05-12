package tree

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/zerodayz7/cmdr/internal/profiles"
)

const IgnoreFileName = ".cmdrignore"

var DefaultIgnoreList = []string{
	".git", "node_modules", "bin", "dist", ".DS_Store",
	"out", "docs", "templates", ".gitattributes", ".gitignore",
	"LICENSE", "README.md", ".vs", ".vscode", "obj", "Assets",
}

// helper do pobierania pełnej ścieżki do pliku ignore
func getIgnorePath() (string, error) {
	dir, err := profiles.GetConfigDir()
	if err != nil {
		return "", err
	}

	// Upewniamy się, że katalog ~/.cmdr istnieje
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(dir, IgnoreFileName), nil
}

func GenerateIgnoreFile() error {
	filePath, err := getIgnorePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filePath); err == nil {
		return nil // Już jest, nie nadpisujemy
	}

	content := strings.Join(DefaultIgnoreList, "\n")
	return os.WriteFile(filePath, []byte(content), 0644)
}

func ReadIgnoreFile() ([]string, error) {
	filePath, err := getIgnorePath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var result []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			result = append(result, line)
		}
	}
	return result, scanner.Err()
}
