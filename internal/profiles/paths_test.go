package profiles

import (
	"os"
	"path/filepath"
	"testing"
)

func setTestHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	return home
}

func TestGetConfigDirUsesUserHome(t *testing.T) {
	home := setTestHome(t)

	dir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir() error: %v", err)
	}

	want := filepath.Join(home, ConfigDirName)
	if dir != want {
		t.Fatalf("GetConfigDir() = %q, want %q", dir, want)
	}
}

func TestEnsureConfigDirCreatesDirectory(t *testing.T) {
	home := setTestHome(t)

	dir, err := EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() error: %v", err)
	}

	want := filepath.Join(home, ConfigDirName)
	if dir != want {
		t.Fatalf("EnsureConfigDir() = %q, want %q", dir, want)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("config directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("config path is not a directory: %q", dir)
	}
}

func TestGetAskConfigPathUsesCentralLocation(t *testing.T) {
	home := setTestHome(t)

	path, err := GetAskConfigPath()
	if err != nil {
		t.Fatalf("GetAskConfigPath() error: %v", err)
	}

	want := filepath.Join(home, ConfigDirName, AskConfigFileName)
	if path != want {
		t.Fatalf("GetAskConfigPath() = %q, want %q", path, want)
	}
	if _, err := os.Stat(filepath.Join(home, ConfigDirName)); err != nil {
		t.Fatalf("config directory was not created: %v", err)
	}
}
