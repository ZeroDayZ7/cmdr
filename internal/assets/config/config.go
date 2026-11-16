package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
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

var supportedFormats = map[string]bool{
	"png":  true,
	"jpg":  true,
	"jpeg": true,
	"gif":  true,
}

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
	return &Config{
		Defaults: Defaults{OutputDir: "generated", Format: "png", Overwrite: false},
		Presets: map[string]Preset{
			"favicon": {
				Sizes:                []int{16, 32, 48, 180, 192, 512},
				Format:               "png",
				PreserveTransparency: true,
			},
		},
	}
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

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON in %s: %w", configPath, err)
	}

	if defaultsRaw, ok := raw["defaults"]; ok {
		var defaults Defaults
		if err := json.Unmarshal(defaultsRaw, &defaults); err != nil {
			return nil, fmt.Errorf("invalid defaults in %s: %w", configPath, err)
		}
		cfg.Defaults = applyDefaults(defaults, cfg.Defaults)
	}

	if presetsRaw, ok := raw["presets"]; ok {
		var presets map[string]Preset
		if err := json.Unmarshal(presetsRaw, &presets); err != nil {
			return nil, fmt.Errorf("invalid presets in %s: %w", configPath, err)
		}
		cfg.Presets = map[string]Preset{}
		for name, preset := range presets {
			filled := applyPresetDefaults(preset, DefaultConfig().Presets["favicon"])
			validated, err := validatePreset(name, filled)
			if err != nil {
				return nil, fmt.Errorf("invalid preset %q in %s: %w", name, configPath, err)
			}
			cfg.Presets[name] = validated
		}
	}
	if _, ok := cfg.Presets["favicon"]; !ok {
		cfg.Presets["favicon"] = DefaultConfig().Presets["favicon"]
	}
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration in %s: %w", configPath, err)
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

	if _, err := os.Stat(configPath); err == nil {
		if !overwrite {
			return fmt.Errorf("configuration already exists at %s; use --force to overwrite", configPath)
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check asset config at %s: %w", configPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("create config directory for %s: %w", configPath, err)
	}
	return writeFileAtomically(configPath, []byte(defaultAssetsConfigJSON), overwrite)
}

func applyDefaults(in Defaults, def Defaults) Defaults {
	if in.OutputDir == "" {
		in.OutputDir = def.OutputDir
	}
	if in.Format == "" {
		in.Format = def.Format
	}
	in.Overwrite = in.Overwrite || def.Overwrite
	return in
}

func applyPresetDefaults(in Preset, def Preset) Preset {
	if in.Format == "" {
		in.Format = def.Format
	}
	if len(in.Sizes) == 0 {
		in.Sizes = append([]int(nil), def.Sizes...)
	}
	if in.PreserveTransparency == false && def.PreserveTransparency {
		in.PreserveTransparency = true
	}
	return in
}

func validateConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	if cfg.Defaults.OutputDir == "" {
		cfg.Defaults.OutputDir = "generated"
	}
	if cfg.Defaults.Format == "" {
		cfg.Defaults.Format = "png"
	}
	cfg.Defaults.Format = strings.ToLower(cfg.Defaults.Format)
	if !supportedFormats[cfg.Defaults.Format] {
		return fmt.Errorf("unsupported default format %q", cfg.Defaults.Format)
	}
	if cfg.Presets == nil {
		cfg.Presets = map[string]Preset{}
	}
	if _, ok := cfg.Presets["favicon"]; !ok {
		cfg.Presets["favicon"] = DefaultConfig().Presets["favicon"]
	}
	for name, preset := range cfg.Presets {
		validated, err := validatePreset(name, preset)
		if err != nil {
			return err
		}
		cfg.Presets[name] = validated
	}
	return nil
}

func validatePreset(name string, preset Preset) (Preset, error) {
	if len(preset.Sizes) == 0 {
		if name == "favicon" {
			preset.Sizes = []int{16, 32, 48, 180, 192, 512}
		} else {
			return Preset{}, fmt.Errorf("preset %q has no sizes", name)
		}
	}
	for _, size := range preset.Sizes {
		if size <= 0 || size > 2048 {
			return Preset{}, fmt.Errorf("preset %q has invalid size %d: expected 1..2048", name, size)
		}
	}
	if preset.Format == "" {
		preset.Format = "png"
	}
	preset.Format = strings.ToLower(preset.Format)
	if !supportedFormats[preset.Format] {
		return Preset{}, fmt.Errorf("preset %q has unsupported format %q", name, preset.Format)
	}
	return preset, nil
}

func writeFileAtomically(path string, data []byte, overwrite bool) error {
	if _, err := os.Stat(path); err == nil && !overwrite {
		return fmt.Errorf("configuration already exists at %s; use --force to overwrite", path)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config directory for %s: %w", path, err)
	}
	tmp, err := os.CreateTemp(dir, "cmdr-assets-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary config file for %s: %w", path, err)
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write temporary config for %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temporary config for %s: %w", path, err)
	}
	if err := replaceFile(tmpPath, path, overwrite); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("replace config at %s: %w", path, err)
	}
	return nil
}

func replaceFile(src, dst string, overwrite bool) error {
	if !overwrite {
		if _, err := os.Stat(dst); err == nil {
			return fmt.Errorf("file already exists: %s", dst)
		}
	}
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := os.Remove(dst); err == nil || os.IsNotExist(err) {
		if err := os.Rename(src, dst); err == nil {
			return nil
		}
	}
	if err := windowsMoveFileEx(src, dst); err == nil {
		return nil
	}
	return fmt.Errorf("replace file %s with %s: %w", src, dst, os.ErrExist)
}

func windowsMoveFileEx(src, dst string) error {
	srcPtr, err := windows.UTF16PtrFromString(src)
	if err != nil {
		return err
	}
	dstPtr, err := windows.UTF16PtrFromString(dst)
	if err != nil {
		return err
	}
	return windows.MoveFileEx(srcPtr, dstPtr, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}
