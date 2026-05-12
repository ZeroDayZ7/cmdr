package config

import (
	"log/slog"
	"os"
	"sync"
)

type Config struct {
	Name    string
	Version string

	LogLevel   slog.Level
	MaxWorkers int
	DryRun     bool

	ProjectRoot string
	Environment string
}

var (
	instance *Config
	once     sync.Once
)

func Get() *Config {
	once.Do(func() {
		instance = &Config{
			Name:        Name,
			Version:     Version,
			LogLevel:    slog.LevelInfo,
			MaxWorkers:  4,
			DryRun:      false,
			Environment: "production",
		}
	})
	return instance
}

func (c *Config) LoadFromEnv() {
	if os.Getenv("CMDR_DEBUG") == "true" {
		c.LogLevel = slog.LevelDebug
	}

	if env := os.Getenv("CMDR_ENV"); env != "" {
		c.Environment = env
	}

	slog.Debug("Config loaded from environment",
		"env", c.Environment,
		"debug", c.LogLevel == slog.LevelDebug)
}

func (c *Config) SetupLogger() {
	opts := &slog.HandlerOptions{
		Level: c.LogLevel,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))
}
