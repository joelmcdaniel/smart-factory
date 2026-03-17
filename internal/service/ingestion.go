package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/joelmcdaniel/go-microservices-and-iot-ingest/smart-factory/internal/core/domain"
	"github.com/joelmcdaniel/go-microservices-and-iot-ingest/smart-factory/internal/core/ports"
)

// IngestionService implements the logic for processing data.
type IngestionService struct {
	repo    ports.SensorRepository
	alerter ports.AlertService
}

// NewIngestionService is the constructor.
// Dependency Injection happens here!
func NewIngestionService(r ports.SensorRepository, a ports.AlertService) *IngestionService {
	return &IngestionService{
		repo:    r,
		alerter: a,
	}
}

func (s *IngestionService) ProcessReading(ctx context.Context, data domain.SensorData) error {
	// 1. Validate (Pure Domain Logic)
	if err := data.Validate(); err != nil {
		slog.ErrorContext(ctx, "Sensor data validation failed",
			slog.String("sensor_id", data.ID),
			slog.String("error", err.Error()),
		)
		return err
	}

	slog.DebugContext(ctx, "Processing sensor reading",
		slog.String("sensor_id", data.ID),
		slog.Float64("temperature", data.Temperature),
		slog.Int64("timestamp", data.Timestamp),
	)

	// 2. Business Logic
	if data.IsCritical() {
		msg := fmt.Sprintf("CRITICAL ALERT: Sensor %s is at %.2f°C", data.ID, data.Temperature)
		slog.WarnContext(ctx, "Critical temperature detected, alerting team",
			slog.String("sensor_id", data.ID),
			slog.Float64("temperature", data.Temperature),
			slog.String("message", msg),
		)
		// We don't know HOW it alerts (Email? Slack? SMS?), we just call the interface
		if err := s.alerter.NotifyTeam(ctx, msg); err != nil {
			slog.ErrorContext(ctx, "Failed to send alert notification",
				slog.String("sensor_id", data.ID),
				slog.String("error", err.Error()),
			)
			return err
		}
	}

	// 3. Persist
	if err := s.repo.Save(ctx, data); err != nil {
		slog.ErrorContext(ctx, "Failed to persist sensor reading",
			slog.String("sensor_id", data.ID),
			slog.String("error", err.Error()),
		)
		return err
	}

	slog.InfoContext(ctx, "Sensor reading processed and persisted successfully",
		slog.String("sensor_id", data.ID),
	)
	return nil
}
