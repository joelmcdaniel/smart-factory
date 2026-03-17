package config

import (
	"log/slog"
	"os"
	"strings"
)

// LogConfig holds logging configuration
type LogConfig struct {
	Level       string `yaml:"level"`       // "debug", "info", "warn", "error"
	ServiceName string `yaml:"serviceName"` // Name of this service
	Format      string `yaml:"format"`      // "json" or "text"
}

// ApplyEnvOverrides applies environment variable overrides to logging config
func (c *LogConfig) ApplyEnvOverrides() {
	// LOG_LEVEL - debug, info, warn, error
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		c.Level = strings.ToLower(level)
	}

	// SERVICE_NAME
	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		c.ServiceName = serviceName
	}

	// LOG_FORMAT - json or text
	if format := os.Getenv("LOG_FORMAT"); format != "" {
		c.Format = strings.ToLower(format)
	}
}

// ParseLogLevel converts string level to slog.Level
func (c *LogConfig) ParseLogLevel() slog.Level {
	switch strings.ToLower(c.Level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// DefaultLogConfig returns sensible defaults for logging configuration
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:       "info",
		ServiceName: "smart-factory",
		Format:      "json",
	}
}
