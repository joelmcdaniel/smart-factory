package mqtt

import (
	"testing"
	"time"

	"github.com/joelmcdaniel/go-microservices-and-iot-ingest/smart-factory/internal/core/domain"
)

// TestParsePayloadJSON validates JSON format parsing
func TestParsePayloadJSON(t *testing.T) {
	tests := []struct {
		name      string
		payload   []byte
		wantError bool
		check     func(sensorData domain.SensorData) bool
	}{
		{
			name:      "valid JSON with all fields",
			payload:   []byte(`{"id":"sensor-01","temperature":23.5,"timestamp":1710585600}`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-01" && sd.Temperature == 23.5 && sd.Timestamp == 1710585600
			},
		},
		{
			name:      "valid JSON without timestamp (uses current time)",
			payload:   []byte(`{"id":"sensor-02","temperature":45.0}`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-02" && sd.Temperature == 45.0 && sd.Timestamp > 0
			},
		},
		{
			name:      "JSON with negative temperature",
			payload:   []byte(`{"id":"sensor-03","temperature":-10.5,"timestamp":1710585600}`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-03" && sd.Temperature == -10.5
			},
		},
		{
			name:      "JSON with zero temperature",
			payload:   []byte(`{"id":"sensor-04","temperature":0.0,"timestamp":1710585600}`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-04" && sd.Temperature == 0.0
			},
		},
		{
			name:      "JSON missing id field",
			payload:   []byte(`{"temperature":23.5,"timestamp":1710585600}`),
			wantError: true,
			check:     func(sd domain.SensorData) bool { return false },
		},
		{
			name:      "invalid JSON syntax",
			payload:   []byte(`{"id":"sensor-01","temperature":23.5`),
			wantError: true,
			check:     func(sd domain.SensorData) bool { return false },
		},
		{
			name:      "JSON with extra fields",
			payload:   []byte(`{"id":"sensor-01","temperature":23.5,"timestamp":1710585600,"extra":"field","ignored":true}`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-01" && sd.Temperature == 23.5
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensorData, err := ParsePayload(tt.payload, string(FormatJSON))
			if (err != nil) != tt.wantError {
				t.Errorf("ParsePayload() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if err == nil && !tt.check(sensorData) {
				t.Errorf("ParsePayload() validation check failed for %+v", sensorData)
			}
		})
	}
}

// TestParsePayloadCSV validates CSV format parsing
func TestParsePayloadCSV(t *testing.T) {
	tests := []struct {
		name      string
		payload   []byte
		wantError bool
		check     func(sensorData domain.SensorData) bool
	}{
		{
			name:      "valid CSV with all fields",
			payload:   []byte(`sensor-01,23.5,1710585600`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-01" && sd.Temperature == 23.5 && sd.Timestamp == 1710585600
			},
		},
		{
			name:      "valid CSV with only id and temperature",
			payload:   []byte(`sensor-02,45.0`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-02" && sd.Temperature == 45.0 && sd.Timestamp > 0
			},
		},
		{
			name:      "CSV with spaces around values",
			payload:   []byte(` sensor-03 , 67.8 , 1710585600 `),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-03" && sd.Temperature == 67.8
			},
		},
		{
			name:      "CSV with negative temperature",
			payload:   []byte(`sensor-04,-15.5,1710585600`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-04" && sd.Temperature == -15.5
			},
		},
		{
			name:      "CSV missing temperature field",
			payload:   []byte(`sensor-05`),
			wantError: true,
			check:     func(sd domain.SensorData) bool { return false },
		},
		{
			name:      "CSV with non-numeric temperature",
			payload:   []byte(`sensor-06,abc,1710585600`),
			wantError: true,
			check:     func(sd domain.SensorData) bool { return false },
		},
		{
			name:      "CSV with invalid timestamp",
			payload:   []byte(`sensor-07,23.5,not-a-number`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				// Invalid timestamp is ignored, current time used
				return sd.ID == "sensor-07" && sd.Temperature == 23.5 && sd.Timestamp > 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensorData, err := ParsePayload(tt.payload, string(FormatCSV))
			if (err != nil) != tt.wantError {
				t.Errorf("ParsePayload() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if err == nil && !tt.check(sensorData) {
				t.Errorf("ParsePayload() validation check failed for %+v", sensorData)
			}
		})
	}
}

// TestParsePayloadRaw validates raw format parsing
func TestParsePayloadRaw(t *testing.T) {
	tests := []struct {
		name      string
		payload   []byte
		wantError bool
		check     func(sensorData domain.SensorData) bool
	}{
		{
			name:      "raw format with just sensor ID",
			payload:   []byte(`sensor-01`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-01" && sd.Temperature == 0 && sd.Timestamp > 0
			},
		},
		{
			name:      "raw format with spaces",
			payload:   []byte(`  sensor-02  `),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-02"
			},
		},
		{
			name:      "raw format with numeric ID",
			payload:   []byte(`12345`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "12345" && sd.Temperature == 0
			},
		},
		{
			name:      "raw space-separated: ID and temperature",
			payload:   []byte(`sensor-03 100.5`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-03" && sd.Temperature == 100.5
			},
		},
		{
			name:      "raw space-separated: ID, temperature, and timestamp",
			payload:   []byte(`sensor-04 45.2 1710585600`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-04" && sd.Temperature == 45.2 && sd.Timestamp == 1710585600
			},
		},
		{
			name:      "raw space-separated: ID with negative temperature",
			payload:   []byte(`sensor-05 -15.5`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-05" && sd.Temperature == -15.5
			},
		},
		{
			name:      "raw space-separated: ID with invalid temperature (ignored, uses 0)",
			payload:   []byte(`sensor-06 not-a-number`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-06" && sd.Temperature == 0
			},
		},
		{
			name:      "raw space-separated: invalid timestamp (ignored)",
			payload:   []byte(`sensor-07 23.5 not-a-timestamp`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-07" && sd.Temperature == 23.5 && sd.Timestamp > 0
			},
		},
		{
			name:      "empty raw payload",
			payload:   []byte(``),
			wantError: true,
			check:     func(sd domain.SensorData) bool { return false },
		},
		{
			name:      "raw payload with only spaces",
			payload:   []byte(`   `),
			wantError: true,
			check:     func(sd domain.SensorData) bool { return false },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensorData, err := ParsePayload(tt.payload, string(FormatRaw))
			if (err != nil) != tt.wantError {
				t.Errorf("ParsePayload() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if err == nil && !tt.check(sensorData) {
				t.Errorf("ParsePayload() validation check failed for %+v", sensorData)
			}
		})
	}
}

// TestAutoDetectFormat validates format auto-detection
func TestAutoDetectFormat(t *testing.T) {
	tests := []struct {
		name        string
		payload     []byte
		expectedFmt PayloadFormat
		wantError   bool
	}{
		{
			name:        "detect JSON format",
			payload:     []byte(`{"id":"sensor-01","temperature":23.5,"timestamp":1710585600}`),
			expectedFmt: FormatJSON,
			wantError:   false,
		},
		{
			name:        "detect CSV format",
			payload:     []byte(`sensor-02,45.0,1710585600`),
			expectedFmt: FormatCSV,
			wantError:   false,
		},
		{
			name:        "fallback to raw format",
			payload:     []byte(`simple-sensor-id`),
			expectedFmt: FormatRaw,
			wantError:   false,
		},
		{
			name:        "JSON without id field falls back",
			payload:     []byte(`{"temperature":23.5,"timestamp":1710585600}`),
			expectedFmt: FormatRaw,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detectedFmt, err := detectFormat(tt.payload)
			if (err != nil) != tt.wantError {
				t.Errorf("detectFormat() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if detectedFmt != tt.expectedFmt {
				t.Errorf("detectFormat() = %v, want %v", detectedFmt, tt.expectedFmt)
			}
		})
	}
}

// TestParsePayloadAutoDetect validates auto-detection within ParsePayload
func TestParsePayloadAutoDetect(t *testing.T) {
	tests := []struct {
		name      string
		payload   []byte
		wantError bool
		check     func(sensorData domain.SensorData) bool
	}{
		{
			name:      "auto-detect JSON",
			payload:   []byte(`{"id":"sensor-01","temperature":23.5,"timestamp":1710585600}`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-01" && sd.Temperature == 23.5
			},
		},
		{
			name:      "auto-detect CSV",
			payload:   []byte(`sensor-02,45.0,1710585600`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-02" && sd.Temperature == 45.0
			},
		},
		{
			name:      "auto-detect raw",
			payload:   []byte(`sensor-03`),
			wantError: false,
			check: func(sd domain.SensorData) bool {
				return sd.ID == "sensor-03" && sd.Temperature == 0
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensorData, err := ParsePayload(tt.payload, string(FormatAuto))
			if (err != nil) != tt.wantError {
				t.Errorf("ParsePayload() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if err == nil && !tt.check(sensorData) {
				t.Errorf("ParsePayload() validation check failed for %+v", sensorData)
			}
		})
	}
}

// TestParsePayloadInvalid tests invalid inputs
func TestParsePayloadInvalid(t *testing.T) {
	tests := []struct {
		name      string
		payload   []byte
		format    string
		wantError bool
	}{
		{
			name:      "invalid format type",
			payload:   []byte(`sensor-01`),
			format:    "unknown",
			wantError: true,
		},
		{
			name:      "empty payload with JSON format",
			payload:   []byte(``),
			format:    string(FormatJSON),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParsePayload(tt.payload, tt.format)
			if (err != nil) != tt.wantError {
				t.Errorf("ParsePayload() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestTimestampHandling validates timestamp defaults and parsing
func TestTimestampHandling(t *testing.T) {
	t.Run("JSON without timestamp uses current time", func(t *testing.T) {
		before := time.Now().Unix()
		payload := []byte(`{"id":"sensor-01","temperature":23.5}`)
		sensorData, err := ParsePayload(payload, string(FormatJSON))
		after := time.Now().Unix()

		if err != nil {
			t.Fatalf("ParsePayload() error = %v", err)
		}
		if sensorData.Timestamp < before || sensorData.Timestamp > after+1 {
			t.Errorf("Timestamp not in expected range: got %d, want between %d and %d", sensorData.Timestamp, before, after+1)
		}
	})

	t.Run("CSV without timestamp uses current time", func(t *testing.T) {
		before := time.Now().Unix()
		payload := []byte(`sensor-02,45.0`)
		sensorData, err := ParsePayload(payload, string(FormatCSV))
		after := time.Now().Unix()

		if err != nil {
			t.Fatalf("ParsePayload() error = %v", err)
		}
		if sensorData.Timestamp < before || sensorData.Timestamp > after+1 {
			t.Errorf("Timestamp not in expected range: got %d, want between %d and %d", sensorData.Timestamp, before, after+1)
		}
	})

	t.Run("explicit timestamp is preserved", func(t *testing.T) {
		payload := []byte(`{"id":"sensor-03","temperature":67.8,"timestamp":1234567890}`)
		sensorData, err := ParsePayload(payload, string(FormatJSON))

		if err != nil {
			t.Fatalf("ParsePayload() error = %v", err)
		}
		if sensorData.Timestamp != 1234567890 {
			t.Errorf("Timestamp = %d, want 1234567890", sensorData.Timestamp)
		}
	})
}

// TestValidationErrors ensures domain validation is enforced
func TestValidationErrors(t *testing.T) {
	t.Run("JSON with empty id fails validation", func(t *testing.T) {
		payload := []byte(`{"id":"","temperature":23.5,"timestamp":1710585600}`)
		_, err := ParsePayload(payload, string(FormatJSON))

		if err == nil {
			t.Errorf("ParsePayload() expected validation error for empty id, got nil")
		}
	})

	t.Run("CSV with empty id fails validation", func(t *testing.T) {
		payload := []byte(`,23.5,1710585600`)
		_, err := ParsePayload(payload, string(FormatCSV))

		if err == nil {
			t.Errorf("ParsePayload() expected validation error for empty id, got nil")
		}
	})

	t.Run("raw with empty id fails validation", func(t *testing.T) {
		_, err := ParsePayload([]byte(``), string(FormatRaw))

		if err == nil {
			t.Errorf("ParsePayload() expected validation error for empty payload, got nil")
		}
	})
}

// TestLargeTemperatureValues validates handling of extreme values
func TestLargeTemperatureValues(t *testing.T) {
	tests := []struct {
		name         string
		payload      []byte
		format       string
		expectedTemp float64
	}{
		{
			name:         "very high temperature (critical alert trigger)",
			payload:      []byte(`{"id":"furnace-01","temperature":1500.5,"timestamp":1710585600}`),
			format:       string(FormatJSON),
			expectedTemp: 1500.5,
		},
		{
			name:         "very low temperature",
			payload:      []byte(`{"id":"freezer-01","temperature":-273.15,"timestamp":1710585600}`),
			format:       string(FormatJSON),
			expectedTemp: -273.15,
		},
		{
			name:         "scientific notation",
			payload:      []byte(`{"id":"sensor-01","temperature":1.23e-4,"timestamp":1710585600}`),
			format:       string(FormatJSON),
			expectedTemp: 0.000123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sensorData, err := ParsePayload(tt.payload, tt.format)
			if err != nil {
				t.Fatalf("ParsePayload() error = %v", err)
			}
			if sensorData.Temperature != tt.expectedTemp {
				t.Errorf("Temperature = %f, want %f", sensorData.Temperature, tt.expectedTemp)
			}
		})
	}
}
