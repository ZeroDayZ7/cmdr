package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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

	if _, err := os.Stat(filePath); err == nil {
		return nil
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %v", filename, err)
	}

	fmt.Printf("Generated %s at %s with default values.\n", filename, exeDir)
	return nil
}

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

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading ignore file: %v", err)
	}

	return result, nil
}
