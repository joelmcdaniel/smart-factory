package service

import (
	"context"
	"fmt"

	"github.com/joelmcdaniel/go-microservices-and-iot-injest/smart-factory/internal/core/domain"
	"github.com/joelmcdaniel/go-microservices-and-iot-injest/smart-factory/internal/core/ports"
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
	// 1. Validate(Pure Domain Logic)
	if err := data.Validate(); err != nil {
		return err
	}

	// 2. Business Logic
	if data.IsCritical() {
		msg := fmt.Sprintf("CRITICAL ALERT: Sensor %s is at %.2f°C", data.ID, data.Temperature)
		// We don't know HOW it alerts (Email? Slack? SMS?), we just call the interface
		if err := s.alerter.NotifyTeam(ctx, msg); err != nil {
			return err //Log this in handler layer
		}
	}

	// 3. Persist
	return s.repo.Save(ctx, data)
}
