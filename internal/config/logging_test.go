package config

import (
	"log/slog"
	"os"
	"testing"
)

func TestDefaultLogConfig(t *testing.T) {
	cfg := DefaultLogConfig()

	if cfg.Level != "info" {
		t.Errorf("default level = %v, want info", cfg.Level)
	}
	if cfg.ServiceName != "smart-factory" {
		t.Errorf("default service name = %v, want smart-factory", cfg.ServiceName)
	}
	if cfg.Format != "json" {
		t.Errorf("default format = %v, want json", cfg.Format)
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  slog.Level
	}{
		{name: "debug", level: "debug", want: slog.LevelDebug},
		{name: "info", level: "info", want: slog.LevelInfo},
		{name: "warn", level: "warn", want: slog.LevelWarn},
		{name: "error", level: "error", want: slog.LevelError},
		{name: "DEBUG uppercase", level: "DEBUG", want: slog.LevelDebug},
		{name: "INFO uppercase", level: "INFO", want: slog.LevelInfo},
		{name: "invalid defaults to info", level: "invalid", want: slog.LevelInfo},
		{name: "empty defaults to info", level: "", want: slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &LogConfig{Level: tt.level}
			got := cfg.ParseLogLevel()
			if got != tt.want {
				t.Errorf("ParseLogLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	t.Run("LOG_LEVEL override", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "debug")
		cfg := DefaultLogConfig()
		cfg.ApplyEnvOverrides()

		if cfg.Level != "debug" {
			t.Errorf("level after env override = %v, want debug", cfg.Level)
		}
	})

	t.Run("SERVICE_NAME override", func(t *testing.T) {
		t.Setenv("SERVICE_NAME", "custom-service")
		cfg := DefaultLogConfig()
		cfg.ApplyEnvOverrides()

		if cfg.ServiceName != "custom-service" {
			t.Errorf("service name after env override = %v, want custom-service", cfg.ServiceName)
		}
	})

	t.Run("LOG_FORMAT override", func(t *testing.T) {
		t.Setenv("LOG_FORMAT", "text")
		cfg := DefaultLogConfig()
		cfg.ApplyEnvOverrides()

		if cfg.Format != "text" {
			t.Errorf("format after env override = %v, want text", cfg.Format)
		}
	})

	t.Run("multiple overrides", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "warn")
		t.Setenv("SERVICE_NAME", "multi-test")
		t.Setenv("LOG_FORMAT", "json")

		cfg := DefaultLogConfig()
		cfg.ApplyEnvOverrides()

		if cfg.Level != "warn" {
			t.Errorf("level = %v, want warn", cfg.Level)
		}
		if cfg.ServiceName != "multi-test" {
			t.Errorf("service name = %v, want multi-test", cfg.ServiceName)
		}
		if cfg.Format != "json" {
			t.Errorf("format = %v, want json", cfg.Format)
		}
	})

	t.Run("case insensitivity for LOG_LEVEL", func(t *testing.T) {
		t.Setenv("LOG_LEVEL", "ERROR")
		cfg := DefaultLogConfig()
		cfg.ApplyEnvOverrides()

		if cfg.Level != "error" {
			t.Errorf("level after case conversion = %v, want error", cfg.Level)
		}
	})

	t.Run("empty env vars don't override defaults", func(t *testing.T) {
		// Ensure env vars are not set
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("SERVICE_NAME")
		os.Unsetenv("LOG_FORMAT")

		cfg := DefaultLogConfig()
		original := *cfg
		cfg.ApplyEnvOverrides()

		if cfg.Level != original.Level {
			t.Errorf("level changed without env var: %v -> %v", original.Level, cfg.Level)
		}
		if cfg.ServiceName != original.ServiceName {
			t.Errorf("service name changed without env var: %v -> %v", original.ServiceName, cfg.ServiceName)
		}
		if cfg.Format != original.Format {
			t.Errorf("format changed without env var: %v -> %v", original.Format, cfg.Format)
		}
	})
}

func TestLogConfigIntegration(t *testing.T) {
	// Test realistic scenario: config with defaults, then override via env
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("SERVICE_NAME", "test-service-integration")

	cfg := DefaultLogConfig()
	cfg.ApplyEnvOverrides()

	level := cfg.ParseLogLevel()

	if cfg.Level != "debug" {
		t.Errorf("final level = %v, want debug", cfg.Level)
	}
	if level != slog.LevelDebug {
		t.Errorf("parsed level = %v, want DEBUG", level)
	}
	if cfg.ServiceName != "test-service-integration" {
		t.Errorf("final service name = %v, want test-service-integration", cfg.ServiceName)
	}
}
