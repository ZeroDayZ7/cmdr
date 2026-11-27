package doctor

import (
	"fmt"
	"io"
)

func WriteText(w io.Writer, report Report, verbose bool) error {
	if _, err := fmt.Fprintln(w, "cmdr doctor"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	section := ""
	for _, result := range report.Results {
		if result.Section != section {
			section = result.Section
			if _, err := fmt.Fprintf(w, "%s\n", section); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(w, "  [%s] %s", result.Status, result.Name); err != nil {
			return err
		}
		if verbose || result.Message != "" {
			if _, err := fmt.Fprintf(w, ": %s", result.Message); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Summary"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  Checks: %d\n", report.Summary.Checks); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  Passed: %d\n", report.Summary.Passed); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  Warnings: %d\n", report.Summary.Warnings); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "  Failed: %d\n", report.Summary.Failed); err != nil {
		return err
	}
	return nil
}
