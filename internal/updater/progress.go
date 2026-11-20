package updater

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

const progressRefreshInterval = 100 * time.Millisecond

// DownloadProgress contains a snapshot of an active download.
type DownloadProgress struct {
	Downloaded int64
	Total      int64
	Elapsed    time.Duration
}

func ProgressPercentage(downloaded, total int64) int {
	if total <= 0 {
		return -1
	}
	if downloaded < 0 {
		return -1
	}
	if downloaded == 0 {
		return 0
	}
	if downloaded >= total {
		return 100
	}
	return int(downloaded * 100 / total)
}

func FormatBytes(bytes int64) string {
	if bytes < 0 {
		return "0 B"
	}
	const unit = float64(1024)
	values := []struct {
		limit float64
		name  string
	}{
		{unit, "KB"},
		{unit * unit, "MB"},
		{unit * unit * unit, "GB"},
	}
	value := float64(bytes)
	if value < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	for i := len(values) - 1; i >= 0; i-- {
		if value >= values[i].limit {
			return fmt.Sprintf("%.1f %s", value/values[i].limit, values[i].name)
		}
	}
	return fmt.Sprintf("%.1f KB", value/unit)
}

func DownloadSpeed(downloaded int64, elapsed time.Duration) float64 {
	if downloaded <= 0 || elapsed <= 0 {
		return 0
	}
	return float64(downloaded) / elapsed.Seconds()
}

func DownloadETA(downloaded, total int64, elapsed time.Duration) time.Duration {
	if downloaded <= 0 || total <= 0 || downloaded >= total {
		return -1
	}
	speed := DownloadSpeed(downloaded, elapsed)
	if speed <= 0 {
		return -1
	}
	remaining := float64(total - downloaded)
	if remaining < 0 {
		return -1
	}
	seconds := remaining / speed
	if seconds < 0 || seconds > float64(24*time.Hour/time.Second) {
		return -1
	}
	return time.Duration(seconds * float64(time.Second))
}

func terminalInteractive(w io.Writer) bool {
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
		return os.Getenv("WT_SESSION") != "" || os.Getenv("ANSICON") != "" || strings.EqualFold(os.Getenv("ConEmuANSI"), "ON") || strings.Contains(strings.ToLower(os.Getenv("TERM")), "ansi") || strings.Contains(strings.ToLower(os.Getenv("TERM")), "xterm")
	}
	term := strings.ToLower(os.Getenv("TERM"))
	return term != "" && term != "dumb"
}

func terminalColor(w io.Writer) bool {
	if _, disabled := os.LookupEnv("NO_COLOR"); disabled {
		return false
	}
	return terminalInteractive(w)
}

// ProgressRenderer renders one download line for interactive terminals and
// sparse status lines for redirected or non-interactive output.
type ProgressRenderer struct {
	w               io.Writer
	interactive     bool
	color           bool
	lastRender      time.Time
	lastInteractive time.Time
}

func NewProgressRenderer(w io.Writer) *ProgressRenderer {
	return &ProgressRenderer{w: w, interactive: terminalInteractive(w), color: terminalColor(w)}
}

func (p *ProgressRenderer) Render(progress DownloadProgress) {
	if p.w == nil {
		return
	}
	now := time.Now()
	if p.interactive {
		if !p.lastRender.IsZero() && now.Sub(p.lastRender) < progressRefreshInterval {
			return
		}
		p.lastRender = now
		percentage := ProgressPercentage(progress.Downloaded, progress.Total)
		bar := progressBar(percentage)
		speed := DownloadSpeed(progress.Downloaded, progress.Elapsed)
		fmt.Fprintf(p.w, "\r\033[K%s", "Downloading ")
		if p.color {
			fmt.Fprintf(p.w, "\033[38;2;163;230;53m")
		}
		fmt.Fprintf(p.w, "%s", bar)
		if p.color {
			fmt.Fprint(p.w, "\033[0m")
		}
		if percentage >= 0 {
			fmt.Fprintf(p.w, "  %3d%%", percentage)
		}
		fmt.Fprintf(p.w, "  %s / %s", FormatBytes(progress.Downloaded), formatProgressTotal(progress.Total))
		fmt.Fprintf(p.w, "  %.1f MB/s", speed/(1024*1024))
		if eta := DownloadETA(progress.Downloaded, progress.Total, progress.Elapsed); eta >= 0 {
			fmt.Fprintf(p.w, "  ETA: %s", formatDuration(eta))
		}
		fmt.Fprintf(p.w, "  elapsed: %s", formatDuration(progress.Elapsed))
		return
	}

	if now.Sub(p.lastInteractive) >= time.Second || p.lastInteractive.IsZero() {
		p.lastInteractive = now
		percentage := ProgressPercentage(progress.Downloaded, progress.Total)
		if percentage >= 0 {
			fmt.Fprintf(p.w, "Downloading: %s / %s (%d%%)\n", FormatBytes(progress.Downloaded), formatProgressTotal(progress.Total), percentage)
		} else {
			fmt.Fprintf(p.w, "Downloading: %s\n", FormatBytes(progress.Downloaded))
		}
	}
}

func (p *ProgressRenderer) Finish(downloaded int64, elapsed time.Duration) {
	if p == nil || p.w == nil {
		return
	}
	if p.interactive {
		fmt.Fprint(p.w, "\r\033[K")
	}
	speed := DownloadSpeed(downloaded, elapsed)
	fmt.Fprintf(p.w, "Downloaded %s in %s (%.1f MB/s)\n", FormatBytes(downloaded), formatDuration(elapsed), speed/(1024*1024))
}

func (p *ProgressRenderer) Abort(reason string) {
	if p == nil || p.w == nil {
		return
	}
	if p.interactive {
		fmt.Fprint(p.w, "\r\033[K")
	}
	fmt.Fprintf(p.w, "Download stopped: %s\n", reason)
}

func progressBar(percentage int) string {
	const width = 24
	if percentage < 0 {
		return "[------------------------]"
	}
	filled := percentage * width / 100
	if filled > width {
		filled = width
	}
	return "[" + strings.Repeat("=", filled) + strings.Repeat("-", width-filled) + "]"
}

func formatProgressTotal(total int64) string {
	if total <= 0 {
		return "unknown"
	}
	return FormatBytes(total)
}

func formatDuration(duration time.Duration) string {
	if duration < 0 {
		return "--:--"
	}
	seconds := int(duration.Round(time.Second).Seconds())
	if seconds < 60 {
		return fmt.Sprintf("00:%02d", seconds)
	}
	minutes := seconds / 60
	seconds %= 60
	if minutes < 60 {
		return fmt.Sprintf("%02d:%02d", minutes, seconds)
	}
	hours := minutes / 60
	minutes %= 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func PrintStage(w io.Writer, name string) {
	green := ""
	reset := ""
	if terminalColor(w) {
		green, reset = "\033[38;2;163;230;53m", "\033[0m"
	}
	fmt.Fprintf(w, "%s[%s]%s %s\n", green, "●", reset, name)
}

func PrintCompleted(w io.Writer, version string) {
	green := ""
	reset := ""
	if terminalColor(w) {
		green, reset = "\033[38;2;163;230;53m", "\033[0m"
	}
	fmt.Fprintf(w, "%s[✓]%s Completed: cmdr updated to v%s.\n", green, reset, version)
}
