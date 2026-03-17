package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// MQTTConfig holds all MQTT client configuration
type MQTTConfig struct {
	Broker               string        `yaml:"broker"`
	ClientID             string        `yaml:"clientId"`
	Topics               []string      `yaml:"topics"`
	KeepAlive            time.Duration `yaml:"keepAlive"`
	ReconnectAttempts    int           `yaml:"reconnectAttempts"`
	ReconnectBackoff     time.Duration `yaml:"reconnectBackoff"`
	ReconnectBackoffMax  time.Duration `yaml:"reconnectBackoffMax"`
	DefaultPayloadFormat string        `yaml:"defaultPayloadFormat"` // "auto", "json", "csv", "raw"
	DisconnectTimeout    time.Duration `yaml:"disconnectTimeout"`
}

// ApplyEnvOverrides applies environment variable overrides to MQTT config
func (c *MQTTConfig) ApplyEnvOverrides() {
	// MQTT_BROKER - required or use default
	if broker := os.Getenv("MQTT_BROKER"); broker != "" {
		c.Broker = broker
	}

	// MQTT_CLIENT_ID
	if clientID := os.Getenv("MQTT_CLIENT_ID"); clientID != "" {
		c.ClientID = clientID
	}

	// MQTT_TOPICS - comma-separated
	if topics := os.Getenv("MQTT_TOPICS"); topics != "" {
		c.Topics = strings.Split(topics, ",")
		// Trim whitespace from each topic
		for i := range c.Topics {
			c.Topics[i] = strings.TrimSpace(c.Topics[i])
		}
	}

	// MQTT_KEEP_ALIVE - in seconds
	if keepAlive := os.Getenv("MQTT_KEEP_ALIVE"); keepAlive != "" {
		if sec, err := strconv.Atoi(keepAlive); err == nil {
			c.KeepAlive = time.Duration(sec) * time.Second
		}
	}

	// MQTT_RECONNECT_ATTEMPTS
	if attempts := os.Getenv("MQTT_RECONNECT_ATTEMPTS"); attempts != "" {
		if n, err := strconv.Atoi(attempts); err == nil {
			c.ReconnectAttempts = n
		}
	}

	// MQTT_RECONNECT_BACKOFF - in seconds
	if backoff := os.Getenv("MQTT_RECONNECT_BACKOFF"); backoff != "" {
		if sec, err := strconv.Atoi(backoff); err == nil {
			c.ReconnectBackoff = time.Duration(sec) * time.Second
		}
	}

	// MQTT_RECONNECT_BACKOFF_MAX - in seconds
	if backoffMax := os.Getenv("MQTT_RECONNECT_BACKOFF_MAX"); backoffMax != "" {
		if sec, err := strconv.Atoi(backoffMax); err == nil {
			c.ReconnectBackoffMax = time.Duration(sec) * time.Second
		}
	}

	// MQTT_DEFAULT_PAYLOAD_FORMAT
	if format := os.Getenv("MQTT_DEFAULT_PAYLOAD_FORMAT"); format != "" {
		c.DefaultPayloadFormat = format
	}

	// MQTT_DISCONNECT_TIMEOUT - in milliseconds
	if timeout := os.Getenv("MQTT_DISCONNECT_TIMEOUT"); timeout != "" {
		if ms, err := strconv.Atoi(timeout); err == nil {
			c.DisconnectTimeout = time.Duration(ms) * time.Millisecond
		}
	}
}

// Validate checks if required fields are set
func (c *MQTTConfig) Validate() error {
	if c.Broker == "" {
		return ErrMissingBroker
	}
	if c.ClientID == "" {
		return ErrMissingClientID
	}
	if len(c.Topics) == 0 {
		return ErrMissingTopics
	}
	return nil
}

// DefaultMQTTConfig returns sensible defaults for MQTT configuration
func DefaultMQTTConfig() *MQTTConfig {
	return &MQTTConfig{
		Broker:               "tcp://localhost:1883",
		ClientID:             "go-factory-listener",
		Topics:               []string{"factory/#"},
		KeepAlive:            60 * time.Second,
		ReconnectAttempts:    5,
		ReconnectBackoff:     1 * time.Second,
		ReconnectBackoffMax:  30 * time.Second,
		DefaultPayloadFormat: "auto",
		DisconnectTimeout:    250 * time.Millisecond,
	}
}
