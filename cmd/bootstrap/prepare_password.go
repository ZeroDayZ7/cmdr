package bootstrap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zerodayz7/cmdr/internal/crypto"
)

var (
	jsonFilePath string
	passLength   int
)

func NewPreparePasswordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "prepare-password",
		Aliases: []string{"prep-pass", "pp"},
		Short:   "Skanuje plik JSON i zastępuje pola 'password' wygenerowanymi hasłami",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonFilePath == "" {
				return fmt.Errorf("podaj ścieżkę do pliku JSON za pomocą flagi -f / --file")
			}

			// 1. Odczyt pliku
			data, err := os.ReadFile(jsonFilePath)
			if err != nil {
				return fmt.Errorf("błąd odczytu pliku: %w", err)
			}

			// 2. Parsowanie do ogólnej structures
			var content interface{}
			if err := json.Unmarshal(data, &content); err != nil {
				return fmt.Errorf("błąd parsowania JSON: %w", err)
			}

			// 3. Rekursywne zastępowanie wartości w JSON
			modified, err := processJSONNode(content, passLength)
			if err != nil {
				return fmt.Errorf("błąd podczas przetwarzania pól: %w", err)
			}

			// 4. Bezpieczny format JSON z wyłączonym eskapowaniem HTML
			var buf bytes.Buffer
			encoder := json.NewEncoder(&buf)
			encoder.SetEscapeHTML(false)
			encoder.SetIndent("", "  ")

			if err := encoder.Encode(modified); err != nil {
				return fmt.Errorf("błąd formatowania JSON: %w", err)
			}

			// 5. Zapis zmodyfikowanego pliku
			if err := os.WriteFile(jsonFilePath, buf.Bytes(), 0644); err != nil {
				return fmt.Errorf("błąd zapisu pliku: %w", err)
			}

			fmt.Printf("✅ Pomyślnie przetworzono plik: %s\n", jsonFilePath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&jsonFilePath, "file", "f", "", "Ścieżka do pliku JSON (wymagane)")
	cmd.Flags().IntVarP(&passLength, "length", "l", 16, "Długość generowanego hasła")

	return cmd
}

// processJSONNode przeszukuje strukturę JSON i zamienia pole "password"
func processJSONNode(node interface{}, length int) (interface{}, error) {
	switch v := node.(type) {
	case map[string]any:
		for key, val := range v {
			if key == "password" {
				// Generujemy nowe bezpieczne hasło z istniejącego modułu internal/crypto
				newPass, err := crypto.GenerateRandomPassword(length)
				if err != nil {
					return nil, err
				}
				v[key] = newPass
			} else {
				var err error
				v[key], err = processJSONNode(val, length)
				if err != nil {
					return nil, err
				}
			}
		}
		return v, nil

	case []any:
		for i, val := range v {
			var err error
			v[i], err = processJSONNode(val, length)
			if err != nil {
				return nil, err
			}
		}
		return v, nil

	default:
		return node, nil
	}
}
