package logger

import (
	"context"
	"testing"
)

func TestGetTraceID(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "trace ID not in context returns fallback",
			ctx:      context.Background(),
			expected: "0000-0000",
		},
		{
			name:     "trace ID in context is retrieved",
			ctx:      WithTraceID(context.Background(), "req-12345"),
			expected: "req-12345",
		},
		{
			name:     "trace ID overwrite works",
			ctx:      WithTraceID(WithTraceID(context.Background(), "old-id"), "new-id"),
			expected: "new-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetTraceID(tt.ctx)
			if got != tt.expected {
				t.Errorf("GetTraceID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWithTraceID(t *testing.T) {
	ctx := context.Background()
	traceID := "trace-abc123"
	newCtx := WithTraceID(ctx, traceID)

	// Verify original context is unchanged
	if GetTraceID(ctx) != "0000-0000" {
		t.Error("original context was modified")
	}

	// Verify new context has the trace ID
	if GetTraceID(newCtx) != traceID {
		t.Errorf("WithTraceID() failed to set trace ID")
	}
}

func TestGetServiceName(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "service name not in context returns empty",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name:     "service name in context is retrieved",
			ctx:      WithServiceName(context.Background(), "smart-factory"),
			expected: "smart-factory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetServiceName(tt.ctx)
			if got != tt.expected {
				t.Errorf("GetServiceName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWithServiceName(t *testing.T) {
	ctx := context.Background()
	serviceName := "my-service"
	newCtx := WithServiceName(ctx, serviceName)

	// Verify original context is unchanged
	if GetServiceName(ctx) != "" {
		t.Error("original context was modified")
	}

	// Verify new context has the service name
	if GetServiceName(newCtx) != serviceName {
		t.Errorf("WithServiceName() failed to set service name")
	}
}

func TestGetRequestID(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		expected string
	}{
		{
			name:     "request ID not in context returns empty",
			ctx:      context.Background(),
			expected: "",
		},
		{
			name:     "request ID in context is retrieved",
			ctx:      WithRequestID(context.Background(), "req-xyz789"),
			expected: "req-xyz789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRequestID(tt.ctx)
			if got != tt.expected {
				t.Errorf("GetRequestID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	requestID := "req-12345"
	newCtx := WithRequestID(ctx, requestID)

	// Verify original context is unchanged
	if GetRequestID(ctx) != "" {
		t.Error("original context was modified")
	}

	// Verify new context has the request ID
	if GetRequestID(newCtx) != requestID {
		t.Errorf("WithRequestID() failed to set request ID")
	}
}

func TestContextChaining(t *testing.T) {
	// Test that multiple context values can be set and retrieved
	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-123")
	ctx = WithServiceName(ctx, "factory-svc")
	ctx = WithRequestID(ctx, "req-456")

	if GetTraceID(ctx) != "trace-123" {
		t.Error("trace ID not preserved in chain")
	}
	if GetServiceName(ctx) != "factory-svc" {
		t.Error("service name not preserved in chain")
	}
	if GetRequestID(ctx) != "req-456" {
		t.Error("request ID not preserved in chain")
	}
}
