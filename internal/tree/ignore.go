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
	// Wersjonowanie i narzędzia programistyczne
	".git", ".github", ".gitattributes", ".gitignore", ".idea", ".vs", ".vscode",
	"LICENSE", "README.md", "scripts", "test",

	// Zbudowane pliki binarne i katalogi wyjściowe
	"bin", "dist", "out", "target", "build", "obj",

	// Web, Node.js i Angular
	"node_modules", ".angular",

	// Flutter i Dart
	".dart_tool", ".flutter-plugins-dependencies", ".metadata",

	// Zasoby i specyficzne dla systemów / IDE
	".DS_Store", "docs", "templates", "raw_assets", "Assets",

	// Platformy mobilne i desktopowe
	"android", "ios", "linux", "macos", "web", "windows",
}

// helper do pobierania pełnej ścieżki do pliku ignore w katalogu konfiguracyjnym
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

	patterns := readIgnoreFileFromPath(filePath)
	return patterns, nil
}

// LoadAllIgnorePatterns agreguje wzorce z pliku konfiguracyjnego, katalogu roboczego oraz obok pliku cmdr.exe
func LoadAllIgnorePatterns() []string {
	var patterns []string

	// 1. Pobierz reguły z globalnego katalogu konfiguracyjnego (~/.cmdr/.cmdrignore)
	if globalPatterns, err := ReadIgnoreFile(); err == nil && globalPatterns != nil {
		patterns = append(patterns, globalPatterns...)
	}

	// 2. Sprawdź plik .cmdrignore obok pliku wykonywalnego cmdr.exe
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		patterns = append(patterns, readIgnoreFileFromPath(filepath.Join(exeDir, IgnoreFileName))...)
	}

	// 3. Sprawdź plik .cmdrignore w bieżącym katalogu roboczym (Working Directory)
	if wd, err := os.Getwd(); err == nil {
		patterns = append(patterns, readIgnoreFileFromPath(filepath.Join(wd, IgnoreFileName))...)
	}

	return uniquePatterns(patterns)
}

// readIgnoreFileFromPath wczytuje i oczyszcza linie z podanej ścieżki pliku
func readIgnoreFileFromPath(fullPath string) []string {
	file, err := os.Open(fullPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Czyszczenie ukośników z końca (np. .git/ staje się .git), aby dopasować do utils.ShouldExclude
		line = strings.TrimSuffix(line, "/")
		line = strings.TrimSuffix(line, "\\")
		lines = append(lines, line)
	}

	// Sprawdzenie błędów skanera eliminuje ostrzeżenie lintera
	if err := scanner.Err(); err != nil {
		// Opcjonalnie: fmt.Fprintln(os.Stderr, "error reading ignore file:", err)
		return lines
	}

	return lines
}

// uniquePatterns usuwa duplikaty reguł z połączonej listy
func uniquePatterns(input []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range input {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
