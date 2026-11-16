package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const defaultAssetsConfigJSON = `{
  "defaults": {
    "output_dir": "generated",
    "format": "png",
    "overwrite": false
  },
  "presets": {
    "favicon": {
      "sizes": [16, 32, 48, 180, 192, 512],
      "format": "png",
      "preserve_transparency": true
    }
  }
}`

type Defaults struct {
	OutputDir string `json:"output_dir"`
	Format    string `json:"format"`
	Overwrite bool   `json:"overwrite"`
}

type Preset struct {
	Sizes                []int  `json:"sizes"`
	Format               string `json:"format"`
	PreserveTransparency bool   `json:"preserve_transparency"`
}

type Config struct {
	Defaults Defaults          `json:"defaults"`
	Presets  map[string]Preset `json:"presets"`
}

func DefaultConfig() *Config {
	cfg := &Config{
		Defaults: Defaults{OutputDir: "generated", Format: "png", Overwrite: false},
		Presets: map[string]Preset{
			"favicon": {
				Sizes:                []int{16, 32, 48, 180, 192, 512},
				Format:               "png",
				PreserveTransparency: true,
			},
		},
	}
	return cfg
}

func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	return filepath.Join(home, ".cmdr"), nil
}

func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "assets.json"), nil
}

func LoadConfig(path string) (*Config, error) {
	configPath := path
	if configPath == "" {
		resolved, err := GetConfigPath()
		if err != nil {
			return nil, err
		}
		configPath = resolved
	}

	cfg := DefaultConfig()
	if _, err := os.Stat(configPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, fmt.Errorf("check asset config at %s: %w", configPath, err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read asset config %s: %w", configPath, err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("invalid JSON in %s: %w", configPath, err)
	}

	if cfg.Defaults.OutputDir == "" {
		cfg.Defaults.OutputDir = "generated"
	}
	if cfg.Defaults.Format == "" {
		cfg.Defaults.Format = "png"
	}
	if cfg.Presets == nil {
		cfg.Presets = map[string]Preset{}
	}
	if _, ok := cfg.Presets["favicon"]; !ok {
		cfg.Presets["favicon"] = DefaultConfig().Presets["favicon"]
	}
	return cfg, nil
}

func InitConfig(path string, overwrite bool) error {
	configPath := path
	if configPath == "" {
		resolved, err := GetConfigPath()
		if err != nil {
			return err
		}
		configPath = resolved
	}

	if _, err := os.Stat(configPath); err == nil && !overwrite {
		return fmt.Errorf("configuration already exists at %s; use --force to overwrite", configPath)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("create config directory for %s: %w", configPath, err)
	}

	if err := os.WriteFile(configPath, []byte(defaultAssetsConfigJSON), 0o644); err != nil {
		return fmt.Errorf("write default config to %s: %w", configPath, err)
	}
	return nil
}
