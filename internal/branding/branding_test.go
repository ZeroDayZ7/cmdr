package branding

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/zerodayz7/cmdr/config"
)

func TestWriteWithoutColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var out bytes.Buffer
	if err := Write(&out); err != nil {
		t.Fatal(err)
	}
	text := out.String()
	if strings.Contains(text, "\033[") {
		t.Fatal("banner contains ANSI escapes when NO_COLOR is set")
	}
	if !strings.Contains(text, " C  cmdr") || !strings.Contains(text, "DEVELOPER TOOLKIT") {
		t.Fatalf("unexpected banner:\n%s", text)
	}
}

func TestSupportsColorRejectsRedirect(t *testing.T) {
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = read.Close() }()
	defer func() { _ = write.Close() }()
	if SupportsColor(write) {
		t.Fatal("pipe should not be detected as a color-capable terminal")
	}
}

func TestWriteIncludesExistingVersion(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	var out bytes.Buffer
	if err := Write(&out); err != nil {
		t.Fatal(err)
	}
	want := "v" + config.Version
	if !strings.Contains(out.String(), want) {
		t.Fatalf("banner does not contain configured version %s: %s", want, out.String())
	}
}
