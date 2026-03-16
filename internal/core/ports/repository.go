package ports

import (
	"context"

	"github.com/joelmcdaniel/go-microservices-and-iot-injest/smart-factory/internal/core/domain"
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
