package tree

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zerodayz7/cmdr/internal/utils"
)

func formatCSV(path string, excludeList []string) (string, error) {
	records := [][]string{{"path", "type"}}
	if err := walkCSV(path, "", excludeList, &records); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	if err := w.WriteAll(records); err != nil {
		return "", fmt.Errorf("failed to write CSV: %w", err)
	}

	return buf.String(), nil
}

func walkCSV(path, prefix string, excludeList []string, records *[][]string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if utils.ShouldExclude(entry.Name(), excludeList) {
			continue
		}

		row := []string{filepath.Join(prefix, entry.Name()), "file"}
		if entry.IsDir() {
			row[1] = "dir"
		}
		*records = append(*records, row)

		if entry.IsDir() {
			err := walkCSV(filepath.Join(path, entry.Name()), filepath.Join(prefix, entry.Name()), excludeList, records)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
