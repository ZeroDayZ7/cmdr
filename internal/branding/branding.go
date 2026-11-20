package branding

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/zerodayz7/cmdr/config"
)

const green = "\033[38;2;163;230;53m"
const light = "\033[97m"
const muted = "\033[90m"
const reset = "\033[0m"

func SupportsColor(w io.Writer) bool {
	if _, disabled := os.LookupEnv("NO_COLOR"); disabled {
		return false
	}

	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	if err != nil || info.Mode()&os.ModeCharDevice == 0 {
		return false
	}

	if runtime.GOOS == "windows" {
		if os.Getenv("WT_SESSION") != "" || os.Getenv("ANSICON") != "" {
			return true
		}
		if strings.EqualFold(os.Getenv("ConEmuANSI"), "ON") {
			return true
		}
		term := strings.ToLower(os.Getenv("TERM"))
		return strings.Contains(term, "xterm") || strings.Contains(term, "ansi")
	}

	term := strings.ToLower(os.Getenv("TERM"))
	return term != "" && term != "dumb"
}

func Write(w io.Writer) {
	name := config.Name
	version := config.Version

	if SupportsColor(w) {
		fmt.Fprintf(w, "  %sC%s  %s%s%s\n", green, reset, light, name, reset)
		fmt.Fprintf(w, "     %sDEVELOPER TOOLKIT · v%s%s\n", muted, version, reset)
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %sAutomate your development workflow.%s\n", muted, reset)
		fmt.Fprintln(w)
		fmt.Fprintf(w, "  %sRun %scmdr --help%s %sto explore commands.%s\n", muted, light, reset, muted, reset)
		return
	}

	fmt.Fprintf(w, "  C  %s\n", name)
	fmt.Fprintf(w, "     DEVELOPER TOOLKIT · v%s\n", version)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  Automate your development workflow.")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  Run cmdr --help to explore commands.")
}
