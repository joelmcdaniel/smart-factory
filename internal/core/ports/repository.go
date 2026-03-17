package ports

import (
	"context"

	"github.com/joelmcdaniel/go-microservices-and-iot-ingest/smart-factory/internal/core/domain"
)

// SensorRepository defines the interface for storage.
// The Core doesn't care if this is Postgres, Mongo, or a text file.
type SensorRepository interface {
	Save(ctx context.Context, data domain.SensorData) error
	GetByID(ctx context.Context, id string) (*domain.SensorData, error)
}

// AlertService defines an interface for sending notifications.
type AlertService interface {
	NotifyTeam(ctx context.Context, message string) error
}

// IngestionService defines the interface for processing sensor readings.
// External adapters (MQTT, HTTP, etc.) depend on this interface, not the concrete implementation.
type IngestionService interface {
	ProcessReading(ctx context.Context, data domain.SensorData) error
}
