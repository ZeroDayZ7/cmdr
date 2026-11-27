package profiles

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	ConfigDirName     = ".cmdr"
	AskConfigFileName = ".config.json"
)

func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	if home == "" {
		return "", fmt.Errorf("resolve user home directory: home directory is empty")
	}

	return filepath.Join(home, ConfigDirName), nil
}

func EnsureConfigDir() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create config directory %s: %w", dir, err)
	}
	return dir, nil
}

func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "profiles.json"), nil
}

func GetAskConfigPath() (string, error) {
	dir, err := EnsureConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, AskConfigFileName), nil
}
