package updater

import (
	"bytes"
	"testing"
	"time"
)

func TestProgressPercentage(t *testing.T) {
	if got := ProgressPercentage(25, 100); got != 25 {
		t.Fatalf("got %d", got)
	}
	if got := ProgressPercentage(0, 0); got != -1 {
		t.Fatalf("unknown = %d", got)
	}
	if got := ProgressPercentage(0, 100); got != 0 {
		t.Fatalf("zero = %d", got)
	}
	if got := ProgressPercentage(101, 100); got != 100 {
		t.Fatalf("clamp = %d", got)
	}
}

func TestFormatBytes(t *testing.T) {
	if got := FormatBytes(512); got != "512 B" {
		t.Fatalf("got %q", got)
	}
	if got := FormatBytes(1024 * 1024); got != "1.0 MB" {
		t.Fatalf("got %q", got)
	}
	if got := FormatBytes(1024 * 1024 * 1024); got != "1.0 GB" {
		t.Fatalf("got %q", got)
	}
}

func TestDownloadSpeedAndETA(t *testing.T) {
	if got := DownloadSpeed(1024*1024, time.Second); got != 1024*1024 {
		t.Fatalf("speed = %f", got)
	}
	eta := DownloadETA(50, 100, time.Second)
	if eta != time.Second {
		t.Fatalf("eta = %v", eta)
	}
	if got := DownloadETA(0, 100, time.Second); got >= 0 {
		t.Fatalf("zero ETA = %v", got)
	}
	if got := DownloadETA(100, 100, time.Second); got >= 0 {
		t.Fatalf("complete ETA = %v", got)
	}
}

func TestProgressRendererNonInteractive(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var out bytes.Buffer
	r := NewProgressRenderer(&out)
	r.Render(DownloadProgress{Downloaded: 25, Total: 100, Elapsed: time.Second})
	if out.Len() == 0 {
		t.Fatal("expected status output")
	}
	if bytes.Contains(out.Bytes(), []byte("\033")) {
		t.Fatal("ANSI escape found")
	}
}

func TestProgressRendererUnknownAndNOColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var out bytes.Buffer
	r := NewProgressRenderer(&out)
	r.Render(DownloadProgress{Downloaded: 0, Total: -1, Elapsed: 0})
	if bytes.Contains(out.Bytes(), []byte("\033")) {
		t.Fatal("ANSI escape found")
	}
}
