package general

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/zerodayz7/cmdr/config"
	"github.com/zerodayz7/cmdr/internal/logger"
)

func executeVersionForTest(args ...string) (string, error) {
	var out bytes.Buffer
	cmd := NewVersionCmd(&logger.ConsoleLogger{})
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestVersionShort(t *testing.T) {
	old := config.Version
	config.Version = "2.1.5"
	defer func() { config.Version = old }()
	out, err := executeVersionForTest("--short")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "2.1.5" {
		t.Fatalf("short version = %q", out)
	}
}

func TestVersionJSON(t *testing.T) {
	out, err := executeVersionForTest("--json")
	if err != nil {
		t.Fatal(err)
	}
	var info config.BuildInfo
	if err := json.Unmarshal([]byte(out), &info); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	if info.Platform == "" || info.GoVersion == "" {
		t.Fatalf("missing runtime fields: %#v", info)
	}
}

func TestVersionUnknownBuildInfo(t *testing.T) {
	oldCommit, oldDate := config.GitCommit, config.BuildDate
	config.GitCommit = ""
	config.BuildDate = ""
	defer func() { config.GitCommit, config.BuildDate = oldCommit, oldDate }()
	out, err := executeVersionForTest()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Git commit: unknown") || !strings.Contains(out, "Build date: unknown") {
		t.Fatalf("unexpected unknown metadata output:\n%s", out)
	}
}
