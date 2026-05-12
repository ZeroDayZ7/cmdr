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

	// 1. Sprawdzamy, czy rozszerzenie jest w ogóle obsługiwane (Profil lub JSON)
	if !isAllowedExtension(ext, cfg) {
		return nil
	}

	// 2. Pobieramy styl komentarza bezpośrednio z mapy w configu
	style, ok := cfg.ProfilesConfig.CommentStyles[ext]
	if !ok {
		return nil
	}

	// 3. Sprawdzamy ignorowane ścieżki (używając pełnej ścieżki, nie tylko nazwy pliku)
	if shouldIgnore(path, cfg) {
		return nil
	}

	// 4. Odczytujemy plik
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read error: %w", err)
	}

	// 5. KRYTYCZNE: Sprawdzamy, czy to nie jest plik binarny (np. exe, png)
	if IsBinary(contentBytes) {
		return nil
	}

	content := string(contentBytes)

	// 6. Sprawdzamy, czy adnotacja już istnieje (limit 10 linii total)
	if HasAnnotation(content) {
		if cfg.Verbose {
			fmt.Printf("[SKIP] %s\n", path)
		}
		return nil
	}

	// 7. Lokalizujemy root projektu (wykorzystując logikę detekcji z profilu)
	root, err := FindProjectRoot(path)
	if err != nil {
		// Jeśli nie znajdziemy roota, używamy katalogu nadrzędnego jako fallback
		root = filepath.Dir(path)
	}

	// 8. Budujemy relatywną ścieżkę do komentarza
	relPath, err := filepath.Rel(root, path)
	if err != nil {
		relPath = filepath.Base(path)
	}

	// Zawsze używamy slashy w komentarzach (Unix-style) nawet na Windows
	comment := fmt.Sprintf(style, filepath.ToSlash(relPath))

	// 9. Generujemy nową zawartość
	newContent := injectComment(content, comment)

	if cfg.DryRun {
		fmt.Printf("[ADD] %s\n", path)
		return nil
	}

	// 10. Zapis
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
