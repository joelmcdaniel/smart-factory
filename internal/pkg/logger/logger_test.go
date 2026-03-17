package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestContextHandlerEnrichesLogs(t *testing.T) {
	// Create a buffer to capture log output
	var buf bytes.Buffer

	// Create logger with context handler
	logger := NewLogger(&buf, slog.LevelInfo)

	ctx := context.Background()
	ctx = WithTraceID(ctx, "test-trace-123")
	ctx = WithServiceName(ctx, "test-service")
	ctx = WithRequestID(ctx, "req-001")

	// Log a message with context
	logger.InfoContext(ctx, "test message", slog.String("custom_field", "value"))

	// Parse the JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to unmarshal JSON log: %v", err)
	}

	// Verify context values were added
	if traceID, ok := logEntry["trace_id"].(string); !ok || traceID != "test-trace-123" {
		t.Errorf("trace_id not found or incorrect: %v", logEntry["trace_id"])
	}
	if service, ok := logEntry["service"].(string); !ok || service != "test-service" {
		t.Errorf("service not found or incorrect: %v", logEntry["service"])
	}
	if requestID, ok := logEntry["request_id"].(string); !ok || requestID != "req-001" {
		t.Errorf("request_id not found or incorrect: %v", logEntry["request_id"])
	}

	// Verify custom field
	if customField, ok := logEntry["custom_field"].(string); !ok || customField != "value" {
		t.Errorf("custom_field not found or incorrect: %v", logEntry["custom_field"])
	}

	// Verify message and level
	if msg, ok := logEntry["msg"].(string); !ok || msg != "test message" {
		t.Errorf("msg not found or incorrect: %v", logEntry["msg"])
	}
	if level, ok := logEntry["level"].(string); !ok || level != "INFO" {
		t.Errorf("level not found or incorrect: %v", logEntry["level"])
	}
}

func TestContextHandlerOmitsFallbackTraceID(t *testing.T) {
	// Test that fallback trace ID "0000-0000" is not added to logs
	var buf bytes.Buffer
	logger := NewLogger(&buf, slog.LevelInfo)

	ctx := context.Background()
	// Don't set trace_id, so it will be "0000-0000"

	logger.InfoContext(ctx, "test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to unmarshal JSON log: %v", err)
	}

	// Verify fallback trace ID is not in the output
	if _, ok := logEntry["trace_id"]; ok {
		t.Error("fallback trace_id should not be added to logs")
	}
}

func TestContextHandlerOmitsEmptyServiceName(t *testing.T) {
	// Test that empty service name is not added to logs
	var buf bytes.Buffer
	logger := NewLogger(&buf, slog.LevelInfo)

	ctx := context.Background()
	// Don't set service name, so it will be empty

	logger.InfoContext(ctx, "test message")

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Fatalf("failed to unmarshal JSON log: %v", err)
	}

	// Verify empty service is not in the output
	if _, ok := logEntry["service"]; ok {
		t.Error("empty service should not be added to logs")
	}
}

func TestContextHandlerWithDifferentLogLevels(t *testing.T) {
	tests := []struct {
		name    string
		level   slog.Level
		logFunc func(*slog.Logger, context.Context, string)
		want    string
	}{
		{
			name:  "info level",
			level: slog.LevelInfo,
			logFunc: func(logger *slog.Logger, ctx context.Context, msg string) {
				logger.InfoContext(ctx, msg)
			},
			want: "INFO",
		},
		{
			name:  "debug level",
			level: slog.LevelDebug,
			logFunc: func(logger *slog.Logger, ctx context.Context, msg string) {
				logger.DebugContext(ctx, msg)
			},
			want: "DEBUG",
		},
		{
			name:  "warn level",
			level: slog.LevelWarn,
			logFunc: func(logger *slog.Logger, ctx context.Context, msg string) {
				logger.WarnContext(ctx, msg)
			},
			want: "WARN",
		},
		{
			name:  "error level",
			level: slog.LevelError,
			logFunc: func(logger *slog.Logger, ctx context.Context, msg string) {
				logger.ErrorContext(ctx, msg)
			},
			want: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, tt.level)

			ctx := WithTraceID(context.Background(), "test-trace")
			tt.logFunc(logger, ctx, "test message")

			var logEntry map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
				t.Fatalf("failed to unmarshal JSON log: %v", err)
			}

			if level, ok := logEntry["level"].(string); !ok || level != tt.want {
				t.Errorf("level = %v, want %v", logEntry["level"], tt.want)
			}

			// Verify trace_id is still present
			if _, ok := logEntry["trace_id"]; !ok {
				t.Error("trace_id should be present in log")
			}
		})
	}
}

func TestNewLoggerCreatesValidLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, slog.LevelInfo)

	if logger == nil {
		t.Error("NewLogger returned nil")
	}

	// Log something to verify it works
	logger.Info("test message")

	if buf.Len() == 0 {
		t.Error("logger did not write to buffer")
	}

	var logEntry map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &logEntry); err != nil {
		t.Errorf("logger output is not valid JSON: %v", err)
	}
}
