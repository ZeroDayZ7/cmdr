package profiles

import (
	"os"
	"path/filepath"
)

func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".cmdr"), nil
}

func GetConfigPath() (string, error) {
	dir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, "profiles.json"), nil
}
