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
	MQTT    *MQTTConfig `yaml:"mqtt"`
	Logging *LogConfig  `yaml:"logging"`
}

// LoadConfig loads configuration from YAML files with environment variable overrides
// Loads both mqtt.yaml and logging.yaml if they exist
func LoadConfig(mqttFilePath, loggingFilePath string) (*Config, error) {
	config := &Config{
		MQTT:    DefaultMQTTConfig(),
		Logging: DefaultLogConfig(),
	}

	// Load MQTT configuration
	if _, err := os.Stat(mqttFilePath); err == nil {
		data, err := os.ReadFile(mqttFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read MQTT config file %s: %w", mqttFilePath, err)
		}

		mqttCfg := &struct {
			MQTT *MQTTConfig `yaml:"mqtt"`
		}{}

		if err := yaml.Unmarshal(data, mqttCfg); err != nil {
			return nil, fmt.Errorf("failed to parse MQTT YAML config: %w", err)
		}

		if mqttCfg.MQTT != nil {
			config.MQTT = mqttCfg.MQTT
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to stat MQTT config file: %w", err)
	}

	// Load Logging configuration
	if _, err := os.Stat(loggingFilePath); err == nil {
		data, err := os.ReadFile(loggingFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read logging config file %s: %w", loggingFilePath, err)
		}

		logCfg := &struct {
			Logging *LogConfig `yaml:"logging"`
		}{}

		if err := yaml.Unmarshal(data, logCfg); err != nil {
			return nil, fmt.Errorf("failed to parse logging YAML config: %w", err)
		}

		if logCfg.Logging != nil {
			config.Logging = logCfg.Logging
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to stat logging config file: %w", err)
	}

	// Apply environment variable overrides
	config.MQTT.ApplyEnvOverrides()
	config.Logging.ApplyEnvOverrides()

	// Validate configuration
	if err := config.MQTT.Validate(); err != nil {
		return nil, fmt.Errorf("invalid MQTT configuration: %w", err)
	}

	return config, nil
}

// LoadConfigOrDefault loads config from files or returns defaults with env overrides
// This is more lenient than LoadConfig - it won't fail if the files are missing
func LoadConfigOrDefault(mqttFilePath, loggingFilePath string) *Config {
	config, err := LoadConfig(mqttFilePath, loggingFilePath)
	if err != nil {
		// Log warning but don't fail startup
		fmt.Fprintf(os.Stderr, "Warning: could not load config from %s or %s: %v, using defaults\n", mqttFilePath, loggingFilePath, err)
		config = &Config{
			MQTT:    DefaultMQTTConfig(),
			Logging: DefaultLogConfig(),
		}
		config.MQTT.ApplyEnvOverrides()
		config.Logging.ApplyEnvOverrides()
	}
	return config
}
