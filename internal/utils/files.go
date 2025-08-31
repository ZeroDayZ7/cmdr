package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GenerateFileWithDefaults zapisuje listę domyślnych wartości do pliku
func GenerateFileWithDefaults(filename string, defaults []string) error {
	if filename == "" {
		filename = ".cmdrignore"
	}
	if defaults == nil {
		defaults = []string{
			".git",
			"node_modules",
			"bin",
			"dist",
			".DS_Store",
			"out",
			"docs",
			"templates",
			".gitattributes",
			".gitignore",
			"LICENSE",
			"README.md",
			".vs",
			".vscode",
			"obj",
			"Assets",
		}
	}

	content := strings.Join(defaults, "\n")

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)

	filePath := filepath.Join(exeDir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err == nil {
		fmt.Printf("Generated %s at %s with default values.\n", filename, exeDir)
	}
	return err
}

// ReadIgnoreFile odczytuje zawartość pliku ignore jako listę
func ReadIgnoreFile(filename string) ([]string, error) {
	if filename == "" {
		filename = ".cmdrignore"
	}

	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)

	filePath := filepath.Join(exeDir, filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var result []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			result = append(result, l)
		}
	}
	return result, nil
}
