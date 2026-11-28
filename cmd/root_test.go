package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func executeForTest(args ...string) (string, error) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	rootCmd.SetArgs(nil)
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	return out.String(), err
}

func TestRootHelpIncludesBanner(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out, err := executeForTest("--help")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "DEVELOPER TOOLKIT") || !strings.Contains(out, "Available Commands:") {
		t.Fatalf("unexpected root help:\n%s", out)
	}
}

func TestSubcommandHelpHasNoBanner(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out, err := executeForTest("assets", "--help")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "DEVELOPER TOOLKIT") {
		t.Fatalf("subcommand help unexpectedly included banner:\n%s", out)
	}
}

func TestVersionCommandHasNoBanner(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out, err := executeForTest("version")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "DEVELOPER TOOLKIT") {
		t.Fatalf("version command unexpectedly included banner:\n%s", out)
	}
}

func TestPDFHelpIncludesExamplesAndNumericPrefix(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	out, err := executeForTest("assets", "pdf", "--help")
	if err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"Examples:", "cmdr assets pdf ./pages", "--recursive --max-size 2MB", "numeric prefix"} {
		if !strings.Contains(out, text) {
			t.Fatalf("PDF help missing %q:\n%s", text, out)
		}
	}
}
