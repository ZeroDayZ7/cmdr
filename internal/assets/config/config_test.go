package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigValidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "assets.json")
	content := `{
		"defaults": {
			"output_dir": "generated",
			"format": "png",
			"overwrite": false
		},
		"presets": {
			"favicon": {
				"sizes": [16, 32, 48],
				"format": "png",
				"preserve_transparency": true
			}
		}
	}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.Defaults.OutputDir != "generated" {
		t.Fatalf("expected output_dir generated, got %q", cfg.Defaults.OutputDir)
	}
	if cfg.Presets["favicon"].Sizes[0] != 16 || cfg.Presets["favicon"].Sizes[1] != 32 {
		t.Fatalf("expected favicon sizes to be loaded, got %#v", cfg.Presets["favicon"].Sizes)
	}
}

func TestLoadConfigMissingFileUsesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not-here.json")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("expected defaults, got error: %v", err)
	}
	if cfg.Defaults.OutputDir == "" {
		t.Fatal("expected default output dir to be set")
	}
	if cfg.Defaults.Format == "" {
		t.Fatal("expected default format to be set")
	}
}

func TestInitConfigCreatesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "assets.json")
	if err := InitConfig(path, false); err != nil {
		t.Fatalf("InitConfig returned error: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist after InitConfig: %v", err)
	}
}

func TestInitConfigRefusesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "assets.json")
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write existing config: %v", err)
	}
	if err := InitConfig(path, false); err == nil {
		t.Fatal("expected existing file to be refused")
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "assets.json")
	if err := os.WriteFile(path, []byte("{bad json"), 0o644); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}
	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected invalid JSON error")
	}
}
