package general

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/zerodayz7/cmdr/internal/profiles"
)

func setAskTestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

func TestWriteAndReadApiConfig(t *testing.T) {
	setAskTestHome(t)

	want := apiConfig{
		"gemini": {
			URL:    "https://example.test/generate",
			ApiKey: "test-api-key",
		},
	}

	path, err := writeApiConfig(want)
	if err != nil {
		t.Fatalf("writeApiConfig() error: %v", err)
	}

	expected, err := profiles.GetAskConfigPath()
	if err != nil {
		t.Fatalf("GetAskConfigPath() error: %v", err)
	}
	if path != expected {
		t.Fatalf("written path = %q, want %q", path, expected)
	}

	got, err := readApiConfig()
	if err != nil {
		t.Fatalf("readApiConfig() error: %v", err)
	}
	if got["gemini"].URL != want["gemini"].URL || got["gemini"].ApiKey != want["gemini"].ApiKey {
		t.Fatalf("readApiConfig() = %#v, want %#v", got, want)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	var onDisk apiConfig
	if err := json.Unmarshal(data, &onDisk); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if onDisk["gemini"].ApiKey != "test-api-key" {
		t.Fatalf("stored API key = %q", onDisk["gemini"].ApiKey)
	}
}

func TestAskConfigIndependentOfWorkingDirectoryAndExecutable(t *testing.T) {
	home := setAskTestHome(t)

	cwd := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(cwd); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldCwd) })

	path, err := writeApiConfig(apiConfig{
		"gemini": {URL: "https://example.test", ApiKey: "key"},
	})
	if err != nil {
		t.Fatalf("writeApiConfig() error: %v", err)
	}

	want := filepath.Join(home, ".cmdr", ".config.json")
	if path != want {
		t.Fatalf("config path = %q, want %q", path, want)
	}
	if _, err := os.Stat(filepath.Join(cwd, ".config.json")); !os.IsNotExist(err) {
		t.Fatalf("unexpected config file in working directory: %v", err)
	}

	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(executable), ".config.json")); !os.IsNotExist(err) {
		t.Fatalf("unexpected config file beside executable: %v", err)
	}
}

func TestReadApiConfigCreatesDefaultInUserConfigDir(t *testing.T) {
	setAskTestHome(t)

	config, err := readApiConfig()
	if err != nil {
		t.Fatalf("readApiConfig() error: %v", err)
	}
	if config["gemini"].ApiKey != "your_gemini_api_key_here" {
		t.Fatalf("unexpected default API key: %q", config["gemini"].ApiKey)
	}

	path, err := profiles.GetAskConfigPath()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("default config was not created at %q: %v", path, err)
	}
}
