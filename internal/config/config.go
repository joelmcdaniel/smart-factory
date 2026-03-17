package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	ErrMissingBroker   = errors.New("MQTT broker URL is required")
	ErrMissingClientID = errors.New("MQTT client ID is required")
	ErrMissingTopics   = errors.New("MQTT topics list is required")
	ErrConfigNotFound  = errors.New("config file not found")
	ErrInvalidYAML     = errors.New("invalid YAML format in config file")
)

// Config is the root configuration struct
type Config struct {
	MQTT *MQTTConfig `yaml:"mqtt"`
}

// LoadConfig loads configuration from a YAML file with environment variable overrides
// If the config file doesn't exist, returns defaults with environment overrides
func LoadConfig(filePath string) (*Config, error) {
	config := &Config{
		MQTT: DefaultMQTTConfig(),
	}

	// Try to load from file if it exists
	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", filePath, err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}

		// Ensure MQTT config is not nil
		if config.MQTT == nil {
			config.MQTT = DefaultMQTTConfig()
		}
	} else if !os.IsNotExist(err) {
		// Return error if stat failed for other reasons (permissions, etc)
		return nil, fmt.Errorf("failed to stat config file: %w", err)
	}
	// If file doesn't exist, continue with defaults (no error)

	// Apply environment variable overrides
	config.MQTT.ApplyEnvOverrides()

	// Validate configuration
	if err := config.MQTT.Validate(); err != nil {
		return nil, fmt.Errorf("invalid MQTT configuration: %w", err)
	}

	return config, nil
}

// LoadConfigOrDefault loads config from file or returns defaults with env overrides
// This is more lenient than LoadConfig - it won't fail if the file is missing
func LoadConfigOrDefault(filePath string) *Config {
	config, err := LoadConfig(filePath)
	if err != nil {
		// Log warning but don't fail startup
		fmt.Fprintf(os.Stderr, "Warning: could not load config from %s: %v, using defaults\n", filePath, err)
		config = &Config{
			MQTT: DefaultMQTTConfig(),
		}
		config.MQTT.ApplyEnvOverrides()
	}
	return config
}
