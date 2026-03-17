package logger

import (
	"context"
	"io"
	"log/slog"
)

// ContextHandler wraps an slog.Handler to automatically inject context values into logs
type ContextHandler struct {
	handler slog.Handler
}

// Handle implements slog.Handler interface, enriching log records with context values
func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	// Extract and add trace_id if present
	if traceID := GetTraceID(ctx); traceID != "" && traceID != "0000-0000" {
		r.AddAttrs(slog.String("trace_id", traceID))
	}

	// Extract and add service name if present
	if serviceName := GetServiceName(ctx); serviceName != "" {
		r.AddAttrs(slog.String("service", serviceName))
	}

	// Extract and add request_id if present
	if requestID := GetRequestID(ctx); requestID != "" {
		r.AddAttrs(slog.String("request_id", requestID))
	}

	// Pass enriched record to underlying handler
	return h.handler.Handle(ctx, r)
}

// Enabled implements slog.Handler interface
func (h ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// WithAttrs implements slog.Handler interface
func (h ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ContextHandler{handler: h.handler.WithAttrs(attrs)}
}

// WithGroup implements slog.Handler interface
func (h ContextHandler) WithGroup(name string) slog.Handler {
	return ContextHandler{handler: h.handler.WithGroup(name)}
}

// NewLogger creates a new structured logger with JSON output and context injection
func NewLogger(w io.Writer, level slog.Level) *slog.Logger {
	baseHandler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(ContextHandler{handler: baseHandler})
}

// SetAsDefault sets the logger as the global default
func SetAsDefault(logger *slog.Logger) {
	slog.SetDefault(logger)
}
