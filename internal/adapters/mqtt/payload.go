package mqtt

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/joelmcdaniel/go-microservices-and-iot-injest/smart-factory/internal/core/domain"
)

// PayloadFormat represents the format of an MQTT message payload
type PayloadFormat string

const (
	FormatAuto PayloadFormat = "auto"
	FormatJSON PayloadFormat = "json"
	FormatCSV  PayloadFormat = "csv"
	FormatRaw  PayloadFormat = "raw"
)

// JSONPayload represents the JSON message structure
type JSONPayload struct {
	ID          string  `json:"id"`
	Temperature float64 `json:"temperature"`
	Timestamp   int64   `json:"timestamp"`
}

// ParsePayload attempts to parse a message payload into domain.SensorData
// It supports JSON, CSV, and raw byte formats with optional auto-detection
func ParsePayload(payload []byte, format string) (domain.SensorData, error) {
	// Determine format
	payloadFormat := PayloadFormat(strings.ToLower(format))

	if payloadFormat == FormatAuto {
		// Auto-detect format
		var err error
		payloadFormat, err = detectFormat(payload)
		if err != nil {
			return domain.SensorData{}, fmt.Errorf("failed to detect payload format: %w", err)
		}
	}

	// Parse based on format
	switch payloadFormat {
	case FormatJSON:
		return parseJSON(payload)
	case FormatCSV:
		return parseCSV(payload)
	case FormatRaw:
		return parseRaw(payload)
	default:
		return domain.SensorData{}, fmt.Errorf("unsupported payload format: %s", format)
	}
}

// parseJSON parses JSON payload format
func parseJSON(payload []byte) (domain.SensorData, error) {
	var jp JSONPayload
	if err := json.Unmarshal(payload, &jp); err != nil {
		return domain.SensorData{}, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	sensorData := domain.SensorData{
		ID:          jp.ID,
		Temperature: jp.Temperature,
		Timestamp:   jp.Timestamp,
	}

	// If timestamp is missing, use current time
	if sensorData.Timestamp == 0 {
		sensorData.Timestamp = time.Now().Unix()
	}

	if err := sensorData.Validate(); err != nil {
		return domain.SensorData{}, fmt.Errorf("JSON payload validation failed: %w", err)
	}

	return sensorData, nil
}

// parseCSV parses CSV payload format (id,temperature,timestamp)
func parseCSV(payload []byte) (domain.SensorData, error) {
	reader := csv.NewReader(strings.NewReader(string(payload)))
	records, err := reader.ReadAll()
	if err != nil {
		return domain.SensorData{}, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) == 0 || len(records[0]) < 2 {
		return domain.SensorData{}, fmt.Errorf("CSV payload must have at least id and temperature fields")
	}

	record := records[0]
	id := strings.TrimSpace(record[0])
	tempStr := strings.TrimSpace(record[1])

	temp, err := strconv.ParseFloat(tempStr, 64)
	if err != nil {
		return domain.SensorData{}, fmt.Errorf("failed to parse temperature value: %w", err)
	}

	timestamp := time.Now().Unix()
	if len(record) > 2 {
		tsStr := strings.TrimSpace(record[2])
		if ts, err := strconv.ParseInt(tsStr, 10, 64); err == nil {
			timestamp = ts
		}
	}

	sensorData := domain.SensorData{
		ID:          id,
		Temperature: temp,
		Timestamp:   timestamp,
	}

	if err := sensorData.Validate(); err != nil {
		return domain.SensorData{}, fmt.Errorf("CSV payload validation failed: %w", err)
	}

	return sensorData, nil
}

// parseRaw attempts to parse raw byte payload with flexible formatting
// Supports:
//   - Plain ID: "sensor-05" → {ID: "sensor-05", Temperature: 0}
//   - Space-separated: "sensor-05 100.5" → {ID: "sensor-05", Temperature: 100.5}
//   - JSON/CSV fallback if parseable
func parseRaw(payload []byte) (domain.SensorData, error) {
	// Try JSON first as a fallback
	var jp JSONPayload
	if err := json.Unmarshal(payload, &jp); err == nil && jp.ID != "" {
		return parseJSON(payload)
	}

	// Try CSV as fallback
	reader := csv.NewReader(strings.NewReader(string(payload)))
	if records, err := reader.ReadAll(); err == nil && len(records) > 0 && len(records[0]) >= 2 {
		return parseCSV(payload)
	}

	// Try space-separated format: "sensor-05 100.5" or "sensor-05 100.5 1710585600"
	parts := strings.Fields(string(payload))
	if len(parts) > 0 {
		id := parts[0]
		temp := 0.0
		timestamp := time.Now().Unix()

		// Try to parse temperature from second field
		if len(parts) > 1 {
			if parsedTemp, err := strconv.ParseFloat(parts[1], 64); err == nil {
				temp = parsedTemp
			}
		}

		// Try to parse timestamp from third field
		if len(parts) > 2 {
			if parsedTS, err := strconv.ParseInt(parts[2], 10, 64); err == nil {
				timestamp = parsedTS
			}
		}

		sensorData := domain.SensorData{
			ID:          id,
			Temperature: temp,
			Timestamp:   timestamp,
		}

		if err := sensorData.Validate(); err != nil {
			return domain.SensorData{}, fmt.Errorf("raw payload validation failed: %w", err)
		}

		return sensorData, nil
	}

	return domain.SensorData{}, fmt.Errorf("raw payload is empty")
}

// detectFormat attempts to auto-detect the payload format
func detectFormat(payload []byte) (PayloadFormat, error) {
	if len(payload) == 0 {
		return "", fmt.Errorf("empty payload")
	}

	// Try JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal(payload, &jsonData); err == nil {
		// Check if it has the expected fields
		if _, hasID := jsonData["id"]; hasID {
			return FormatJSON, nil
		}
	}

	// Try CSV
	reader := csv.NewReader(strings.NewReader(string(payload)))
	if records, err := reader.ReadAll(); err == nil && len(records) > 0 && len(records[0]) >= 2 {
		// Check if second field is a valid number (temperature)
		if _, err := strconv.ParseFloat(records[0][1], 64); err == nil {
			return FormatCSV, nil
		}
	}

	// Default to raw
	return FormatRaw, nil
}
