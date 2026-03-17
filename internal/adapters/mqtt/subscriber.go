package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/joelmcdaniel/go-microservices-and-iot-ingest/smart-factory/internal/config"
	"github.com/joelmcdaniel/go-microservices-and-iot-ingest/smart-factory/internal/core/ports"
	"github.com/joelmcdaniel/go-microservices-and-iot-ingest/smart-factory/internal/pkg/logger"
)

// Subscriber manages MQTT client connection and message handling
type Subscriber struct {
	client             mqtt.Client
	config             *config.MQTTConfig
	service            ports.IngestionService
	mu                 sync.RWMutex
	connected          bool
	stopReconnect      chan struct{}
	messageProcessorCh chan MessageToProcess
	stopProcessing     chan struct{}
	processingWg       sync.WaitGroup
	metrics            SubscriberMetrics
}

// SubscriberMetrics holds operational metrics with thread-safe access
type SubscriberMetrics struct {
	MessagesReceived  int64
	MessagesProcessed int64
	ProcessErrors     int64
	mu                sync.RWMutex
}

// SubscriberMetricsSnapshot is a safe, copy-friendly snapshot of metrics (no mutex)
type SubscriberMetricsSnapshot struct {
	MessagesReceived  int64
	MessagesProcessed int64
	ProcessErrors     int64
}

// MessageToProcess represents a message queued for processing
type MessageToProcess struct {
	Topic    string
	Payload  []byte
	Received time.Time
}

// NewSubscriber creates a new MQTT subscriber
func NewSubscriber(cfg *config.MQTTConfig, svc ports.IngestionService) *Subscriber {
	return &Subscriber{
		config:             cfg,
		service:            svc,
		stopReconnect:      make(chan struct{}),
		messageProcessorCh: make(chan MessageToProcess, 100), // Buffered channel for async processing
		stopProcessing:     make(chan struct{}),
		metrics:            SubscriberMetrics{},
	}
}

// Connect establishes connection to the MQTT broker
func (s *Subscriber) Connect(ctx context.Context) error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(s.config.Broker)
	opts.SetClientID(s.config.ClientID)
	opts.SetKeepAlive(s.config.KeepAlive)
	opts.SetAutoReconnect(false) // We handle reconnection manually
	opts.SetDefaultPublishHandler(s.defaultMessageHandler())

	// Set up connection callbacks
	opts.OnConnect = s.onConnect
	opts.OnConnectionLost = s.onConnectionLost

	s.client = mqtt.NewClient(opts)

	// Connect synchronously
	token := s.client.Connect()
	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("timeout connecting to MQTT broker: %s", s.config.Broker)
	}
	if token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	s.mu.Lock()
	s.connected = true
	s.mu.Unlock()

	slog.Info("Connected to MQTT broker",
		slog.String("broker", s.config.Broker),
	)

	// Start message processor goroutine
	s.processingWg.Add(1)
	go s.messageProcessor()

	return nil
}

// Subscribe subscribes to configured topics
func (s *Subscriber) Subscribe() error {
	if !s.isConnected() {
		return fmt.Errorf("not connected to MQTT broker")
	}

	for _, topic := range s.config.Topics {
		token := s.client.Subscribe(topic, 0, s.messageHandler)
		if !token.WaitTimeout(10 * time.Second) {
			return fmt.Errorf("timeout subscribing to topic: %s", topic)
		}
		if token.Error() != nil {
			return fmt.Errorf("failed to subscribe to topic %s: %w", topic, token.Error())
		}
		slog.Debug("Subscribed to MQTT topic",
			slog.String("topic", topic),
		)
	}

	return nil
}

// Disconnect gracefully closes the MQTT connection and stops processing
func (s *Subscriber) Disconnect() {
	slog.Info("Disconnecting from MQTT broker")

	// Stop reconnection attempts
	close(s.stopReconnect)

	// Stop message processing
	close(s.stopProcessing)
	s.processingWg.Wait()

	// Disconnect from broker
	if s.client != nil && s.isConnected() {
		s.client.Disconnect(uint(s.config.DisconnectTimeout.Milliseconds()))
	}

	s.mu.Lock()
	s.connected = false
	s.mu.Unlock()

	slog.Info("Disconnected from MQTT broker")
}

// isConnected checks if the client is connected
func (s *Subscriber) isConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.client != nil && s.client.IsConnected()
}

// GetMetrics returns a snapshot of current subscriber metrics (safe to copy)
func (s *Subscriber) GetMetrics() SubscriberMetricsSnapshot {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()
	return SubscriberMetricsSnapshot{
		MessagesReceived:  s.metrics.MessagesReceived,
		MessagesProcessed: s.metrics.MessagesProcessed,
		ProcessErrors:     s.metrics.ProcessErrors,
	}
}

// defaultMessageHandler returns the default message handler for all topics
func (s *Subscriber) defaultMessageHandler() mqtt.MessageHandler {
	return func(client mqtt.Client, msg mqtt.Message) {
		s.messageHandler(client, msg)
	}
}

// onConnect is called when the client connects
func (s *Subscriber) onConnect(client mqtt.Client) {
	s.mu.Lock()
	s.connected = true
	s.mu.Unlock()
	slog.Info("MQTT client connected to broker")
}

// onConnectionLost is called when the connection is lost
func (s *Subscriber) onConnectionLost(client mqtt.Client, err error) {
	s.mu.Lock()
	s.connected = false
	s.mu.Unlock()
	slog.Warn("MQTT connection lost, starting reconnection attempts",
		slog.String("error", err.Error()),
	)

	// Start reconnection loop
	go s.reconnectLoop()
}

// reconnectLoop manages exponential backoff reconnection
func (s *Subscriber) reconnectLoop() {
	backoff := s.config.ReconnectBackoff
	maxBackoff := s.config.ReconnectBackoffMax

	for attempt := 0; attempt < s.config.ReconnectAttempts; attempt++ {
		select {
		case <-s.stopReconnect:
			slog.Info("Reconnection loop stopped")
			return
		case <-time.After(backoff):
			slog.Debug("MQTT reconnection attempt",
				slog.Int("attempt", attempt+1),
				slog.Int("max_attempts", s.config.ReconnectAttempts),
				slog.Duration("backoff", backoff),
			)

			if err := s.reconnect(); err != nil {
				slog.Debug("Reconnection failed", slog.String("error", err.Error()))
			} else {
				// Reconnection succeeded, try to resubscribe
				if err := s.Subscribe(); err != nil {
					slog.Error("Failed to resubscribe after reconnection",
						slog.String("error", err.Error()),
					)
				} else {
					// Successfully reconnected and resubscribed
					slog.Info("Successfully reconnected to MQTT broker")
					return
				}
			}

			// Increase backoff for next attempt
			backoff = time.Duration(float64(backoff) * 1.5)
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}

	slog.Error("Failed to reconnect to MQTT broker",
		slog.Int("attempts", s.config.ReconnectAttempts),
	)
}

// reconnect attempts to reconnect to the broker
func (s *Subscriber) reconnect() error {
	if s.client == nil {
		return fmt.Errorf("MQTT client not initialized")
	}

	token := s.client.Connect()
	if !token.WaitTimeout(10 * time.Second) {
		return fmt.Errorf("timeout reconnecting to broker")
	}
	return token.Error()
}

// messageHandler queues messages for async processing
func (s *Subscriber) messageHandler(client mqtt.Client, msg mqtt.Message) {
	s.metrics.mu.Lock()
	s.metrics.MessagesReceived++
	s.metrics.mu.Unlock()

	select {
	case s.messageProcessorCh <- MessageToProcess{
		Topic:    msg.Topic(),
		Payload:  msg.Payload(),
		Received: time.Now(),
	}:
	case <-s.stopProcessing:
		slog.Debug("Message processor stopped, discarding message",
			slog.String("topic", msg.Topic()),
		)
	}
}

// messageProcessor handles processing of queued messages
func (s *Subscriber) messageProcessor() {
	defer s.processingWg.Done()

	for {
		select {
		case <-s.stopProcessing:
			slog.Info("Message processor shutting down")
			return
		case msg := <-s.messageProcessorCh:
			s.processMessage(msg)
		}
	}
}

// processMessage processes a single message
func (s *Subscriber) processMessage(msg MessageToProcess) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a trace ID based on message topic and timestamp
	traceID := fmt.Sprintf("%s-%d", msg.Topic, msg.Received.UnixNano())
	ctx = logger.WithTraceID(ctx, traceID)

	// Parse payload and invoke service
	sensorData, err := ParsePayload(msg.Payload, s.config.DefaultPayloadFormat)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to parse MQTT payload",
			slog.String("topic", msg.Topic),
			slog.String("error", err.Error()),
		)
		s.metrics.mu.Lock()
		s.metrics.ProcessErrors++
		s.metrics.mu.Unlock()
		return
	}

	// Process the reading through the service
	if err := s.service.ProcessReading(ctx, sensorData); err != nil {
		slog.ErrorContext(ctx, "Failed to process sensor reading",
			slog.String("sensor_id", sensorData.ID),
			slog.String("error", err.Error()),
		)
		s.metrics.mu.Lock()
		s.metrics.ProcessErrors++
		s.metrics.mu.Unlock()
		return
	}

	s.metrics.mu.Lock()
	s.metrics.MessagesProcessed++
	s.metrics.mu.Unlock()

	slog.InfoContext(ctx, "Successfully processed sensor reading",
		slog.String("sensor_id", sensorData.ID),
		slog.Float64("temperature", sensorData.Temperature),
	)
}
