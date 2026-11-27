package general

import (
	"bytes"
	"strings"
	"testing"
)

func executeAskForTest(args ...string) (string, error) {
	var out bytes.Buffer
	cmd := NewAskCmd()
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestAskHelpDocumentsModernCommands(t *testing.T) {
	out, err := executeAskForTest("--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"Interactions", "--model", "models", "config"} {
		if !strings.Contains(out, text) {
			t.Fatalf("ask help missing %q:\n%s", text, out)
		}
	}
}

func TestAskConfigSubcommandDoesNotPrintAPIKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("GEMINI_API_KEY", "secret-test-key")

	out, err := executeAskForTest("config")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "secret-test-key") {
		t.Fatalf("configuration output exposed API key: %s", out)
	}
	if !strings.Contains(out, "API key:") {
		t.Fatalf("configuration status missing: %s", out)
	}
}
