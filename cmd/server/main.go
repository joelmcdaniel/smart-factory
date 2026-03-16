package main

import (
	"context"
	"log"

	"github.com/joelmcdaniel/go-microservices-and-iot-injest/smart-factory/internal/adapters/repo"
	"github.com/joelmcdaniel/go-microservices-and-iot-injest/smart-factory/internal/core/domain"
	"github.com/joelmcdaniel/go-microservices-and-iot-injest/smart-factory/internal/service"
)

// MockAlert is a simple adapter for the AlertService port
type MockAlert struct{}

func (m *MockAlert) NotifyTeam(ctx context.Context, msg string) error {
	log.Println("[MOCK EMAIL SENT]:", msg)
	return nil
}

func main() {
	// 1. Initialize Adapters (Infrastructure)
	// In production, we would swap 'NewMemoryRepo' for 'NewPostgresRepo'
	sensorRepo := repo.NewMemoryRepo()
	alertAdapter := &MockAlert{}

	// 2. Initialize Service (Core)
	// Inject the adapters into the service
	ingestionSvc := service.NewIngestionService(sensorRepo, alertAdapter)

	// 3. Run the Application
	ctx := context.Background()
	reading := domain.SensorData{
		ID:          "turbine-X1",
		Temperature: 105.5, // This is critical!
		Timestamp:   1678900000,
	}

	log.Println("Starting Ingestion...")
	if err := ingestionSvc.ProcessReading(ctx, reading); err != nil {
		log.Fatal("Failed to process:", err)
	}
	log.Println("Ingestion Complete.")
}
