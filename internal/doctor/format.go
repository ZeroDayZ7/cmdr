package doctor

import (
	"fmt"
	"io"
)

func WriteText(w io.Writer, report Report, verbose bool) {
	fmt.Fprintln(w, "cmdr doctor")
	fmt.Fprintln(w)
	section := ""
	for _, result := range report.Results {
		if result.Section != section {
			section = result.Section
			fmt.Fprintf(w, "%s\n", section)
		}
		fmt.Fprintf(w, "  [%s] %s", result.Status, result.Name)
		if verbose || result.Message != "" {
			fmt.Fprintf(w, ": %s", result.Message)
		}
		fmt.Fprintln(w)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Summary")
	fmt.Fprintf(w, "  Checks: %d\n", report.Summary.Checks)
	fmt.Fprintf(w, "  Passed: %d\n", report.Summary.Passed)
	fmt.Fprintf(w, "  Warnings: %d\n", report.Summary.Warnings)
	fmt.Fprintf(w, "  Failed: %d\n", report.Summary.Failed)
}
