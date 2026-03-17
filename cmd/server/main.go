package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joelmcdaniel/go-microservices-and-iot-injest/smart-factory/internal/adapters/mqtt"
	"github.com/joelmcdaniel/go-microservices-and-iot-injest/smart-factory/internal/adapters/repo"
	"github.com/joelmcdaniel/go-microservices-and-iot-injest/smart-factory/internal/config"
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
	log.Println("=== Smart Factory Server Startup ===")

	// 1. Load Configuration (YAML + environment variable overrides)
	log.Println("[CONFIG] Loading MQTT configuration from config/mqtt.yaml...")
	cfg, err := config.LoadConfig("./config/mqtt.yaml")
	if err != nil {
		log.Printf("[CONFIG] Warning: %v, using defaults", err)
		cfg = &config.Config{
			MQTT: config.DefaultMQTTConfig(),
		}
		cfg.MQTT.ApplyEnvOverrides()
	}
	log.Printf("[CONFIG] MQTT Broker: %s", cfg.MQTT.Broker)
	log.Printf("[CONFIG] Topics: %v", cfg.MQTT.Topics)

	// 2. Initialize Adapters (Infrastructure)
	log.Println("[INIT] Initializing adapters...")
	sensorRepo := repo.NewMemoryRepo()
	alertAdapter := &MockAlert{}
	log.Println("[INIT] ✓ Repository adapter initialized (in-memory)")
	log.Println("[INIT] ✓ Alert adapter initialized (mock)")

	// 3. Initialize Service (Core)
	log.Println("[INIT] Initializing core service...")
	ingestionSvc := service.NewIngestionService(sensorRepo, alertAdapter)
	log.Println("[INIT] ✓ Ingestion service initialized")

	// 4. Initialize MQTT Subscriber (Inbound Adapter)
	log.Println("[INIT] Initializing MQTT subscriber...")
	subscriber := mqtt.NewSubscriber(cfg.MQTT, ingestionSvc)

	// Connect to MQTT broker
	ctx := context.Background()
	if err := subscriber.Connect(ctx); err != nil {
		log.Fatalf("[MQTT] Failed to connect to broker: %v", err)
	}

	// Subscribe to configured topics
	if err := subscriber.Subscribe(); err != nil {
		log.Fatalf("[MQTT] Failed to subscribe to topics: %v", err)
	}
	log.Println("[INIT] ✓ MQTT subscriber initialized and connected")

	// 5. Set up graceful shutdown on interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 6. Demo: Process a test reading (optional - can be removed/commented out)
	demoReading()

	log.Println("=== Server Running ===")
	log.Println("[INFO] Press Ctrl+C to shutdown...")

	// Block until shutdown signal
	<-sigChan
	log.Println("\n[SHUTDOWN] Received interrupt signal, shutting down gracefully...")

	// 7. Graceful shutdown
	subscriber.Disconnect()
	log.Println("[SHUTDOWN] MQTT subscriber disconnected")

	log.Println("=== Server Shutdown Complete ===")
}

// demoReading demonstrates a test reading (optional for demo purposes)
func demoReading() {
	log.Println("\n=== Demo: Processing Test Reading ===")
	ctx := context.Background()
	alertAdapter := &MockAlert{}
	sensorRepo := repo.NewMemoryRepo()
	ingestionSvc := service.NewIngestionService(sensorRepo, alertAdapter)

	reading := domain.SensorData{
		ID:          "turbine-X1",
		Temperature: 105.5, // This is critical!
		Timestamp:   1678900000,
	}

	log.Println("[DEMO] Processing reading...")
	if err := ingestionSvc.ProcessReading(ctx, reading); err != nil {
		log.Printf("[DEMO] Error: %v", err)
	}
	log.Println("[DEMO] ✓ Demo reading processed")
}
