package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joelmcdaniel/go-microservices-and-iot-ingest/smart-factory/internal/adapters/mqtt"
	"github.com/joelmcdaniel/go-microservices-and-iot-ingest/smart-factory/internal/adapters/repo"
	"github.com/joelmcdaniel/go-microservices-and-iot-ingest/smart-factory/internal/config"
	"github.com/joelmcdaniel/go-microservices-and-iot-ingest/smart-factory/internal/core/domain"
	"github.com/joelmcdaniel/go-microservices-and-iot-ingest/smart-factory/internal/pkg/logger"
	"github.com/joelmcdaniel/go-microservices-and-iot-ingest/smart-factory/internal/service"
)

// MockAlert is a simple adapter for the AlertService port
type MockAlert struct{}

func (m *MockAlert) NotifyTeam(ctx context.Context, msg string) error {
	slog.InfoContext(ctx, "Mock alert sent",
		slog.String("message", msg),
	)
	return nil
}

func main() {
	// 0. Initialize structured logger (FIRST - before config loading)
	logCfg := config.DefaultLogConfig()
	logCfg.ApplyEnvOverrides()
	logLevel := logCfg.ParseLogLevel()

	structuredLogger := logger.NewLogger(os.Stdout, logLevel)
	logger.SetAsDefault(structuredLogger)

	// Create context with service name for all logs
	ctx := context.Background()
	ctx = logger.WithServiceName(ctx, logCfg.ServiceName)

	slog.InfoContext(ctx, "Starting Smart Factory Server")

	// 1. Load Configuration (YAML + environment variable overrides)
	slog.InfoContext(ctx, "Loading configuration from config files")
	cfg, err := config.LoadConfig("./config/mqtt.yaml", "./config/logging.yaml")
	if err != nil {
		slog.WarnContext(ctx, "Failed to load config, using defaults",
			slog.String("error", err.Error()),
		)
		cfg = &config.Config{
			MQTT:    config.DefaultMQTTConfig(),
			Logging: config.DefaultLogConfig(),
		}
		cfg.MQTT.ApplyEnvOverrides()
		cfg.Logging.ApplyEnvOverrides()
	}

	slog.InfoContext(ctx, "Configuration loaded",
		slog.String("mqtt_broker", cfg.MQTT.Broker),
		slog.Any("mqtt_topics", cfg.MQTT.Topics),
	)

	// 2. Initialize Adapters (Infrastructure)
	slog.InfoContext(ctx, "Initializing adapters")
	sensorRepo := repo.NewMemoryRepo()
	alertAdapter := &MockAlert{}
	slog.DebugContext(ctx, "Repository adapter initialized", slog.String("type", "in-memory"))
	slog.DebugContext(ctx, "Alert adapter initialized", slog.String("type", "mock"))

	// 3. Initialize Service (Core)
	slog.InfoContext(ctx, "Initializing core service")
	ingestionSvc := service.NewIngestionService(sensorRepo, alertAdapter)
	slog.DebugContext(ctx, "Ingestion service initialized")

	// 4. Initialize MQTT Subscriber (Inbound Adapter)
	slog.InfoContext(ctx, "Initializing MQTT subscriber")
	subscriber := mqtt.NewSubscriber(cfg.MQTT, ingestionSvc)

	// Connect to MQTT broker
	if err := subscriber.Connect(ctx); err != nil {
		slog.ErrorContext(ctx, "Failed to connect to MQTT broker",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	// Subscribe to configured topics
	if err := subscriber.Subscribe(); err != nil {
		slog.ErrorContext(ctx, "Failed to subscribe to topics",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}
	slog.InfoContext(ctx, "MQTT subscriber initialized and connected")

	// 5. Set up graceful shutdown on interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 6. Demo: Process a test reading (optional - can be removed/commented out)
	demoReading(ctx)

	slog.InfoContext(ctx, "Server running, press Ctrl+C to shutdown")

	// Block until shutdown signal
	<-sigChan
	slog.WarnContext(ctx, "Received interrupt signal, shutting down gracefully")

	// 7. Graceful shutdown
	subscriber.Disconnect()
	slog.InfoContext(ctx, "MQTT subscriber disconnected")

	slog.InfoContext(ctx, "Server shutdown complete")
}

// demoReading demonstrates a test reading (optional for demo purposes)
func demoReading(ctx context.Context) {
	slog.InfoContext(ctx, "Processing demo test reading")
	alertAdapter := &MockAlert{}
	sensorRepo := repo.NewMemoryRepo()
	ingestionSvc := service.NewIngestionService(sensorRepo, alertAdapter)

	reading := domain.SensorData{
		ID:          "turbine-X1",
		Temperature: 105.5, // This is critical!
		Timestamp:   1678900000,
	}

	slog.InfoContext(ctx, "Processing reading",
		slog.String("sensor_id", reading.ID),
		slog.Float64("temperature", reading.Temperature),
	)
	if err := ingestionSvc.ProcessReading(ctx, reading); err != nil {
		slog.ErrorContext(ctx, "Error processing demo reading",
			slog.String("error", err.Error()),
		)
		return
	}
	slog.InfoContext(ctx, "Demo reading processed successfully")
}
