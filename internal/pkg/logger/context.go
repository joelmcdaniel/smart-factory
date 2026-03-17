package logger

import "context"

// contextKey is a custom type for context keys to prevent collisions
type contextKey string

const (
	traceIDKey   contextKey = "trace_id"
	serviceKey   contextKey = "service"
	requestIDKey contextKey = "request_id"
)

// GetTraceID retrieves the trace ID from context, returns fallback if not found
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return "0000-0000"
}

// WithTraceID adds a trace ID to the context
func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey, id)
}

// GetServiceName retrieves the service name from context
func GetServiceName(ctx context.Context) string {
	if name, ok := ctx.Value(serviceKey).(string); ok {
		return name
	}
	return ""
}

// WithServiceName adds a service name to the context
func WithServiceName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, serviceKey, name)
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}
