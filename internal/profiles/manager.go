package profiles

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

func LoadConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return initDefaultConfig()
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not read config from %s: %w", configPath, err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid json in %s: %w", configPath, err)
	}

	sort.Slice(cfg.Profiles, func(i, j int) bool {
		return cfg.Profiles[i].Priority > cfg.Profiles[j].Priority
	})

	return &cfg, nil
}

func initDefaultConfig() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(configPath, []byte(DefaultProfilesConfig), 0644); err != nil {
		return nil, fmt.Errorf("failed to create default config at %s: %w", configPath, err)
	}

	fmt.Printf("⚙️  Initialized default profiles at: %s\n", configPath)

	var cfg Config
	_ = json.Unmarshal([]byte(DefaultProfilesConfig), &cfg)
	return &cfg, nil
}
