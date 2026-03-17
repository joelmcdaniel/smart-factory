## MQTT Subscriber Integration for Smart Factory

### Architecture Overview

The MQTT subscriber has been integrated as an **inbound adapter** in the hexagonal architecture:

```
MQTT Broker (external)
    ↓ (subscribe to factory/# topics)
internal/adapters/mqtt/
    ├── subscriber.go       (main orchestrator - connection, message processing)
    ├── payload.go          (multi-format message parsing)
    ├── handler.go          (reserved for future enhancements)
    ├── connection.go       (reserved for future enhancements)
    └── reconnect.go        (reserved for future enhancements)
    
    ↓ (service injection)
    
internal/core/
    ├── domain/             (pure business entities)
    │   └── sensor.go       (SensorData validation & business rules)
    ├── ports/              (interface contracts)
    │   └── repository.go   (SensorRepository, AlertService, IngestionService)
    └── service/            (unchanged!)
        └── ingestion.go    (ProcessReading accepts SensorData from any source)

    ↓ (through ports)
    
internal/adapters/
    ├── repo/               (data persistence)
    │   └── memory.go       (in-memory implementation)
    └── mqtt/               (message ingestion)
        └── subscriber.go   (MQTT message source)
```

**Key architectural principle**: The service layer (`IngestionService`) **doesn't know about MQTT**. It simply processes `SensorData` objects received through the `ProcessReading()` method. This allows future adapters (HTTP, gRPC, file upload, Kafka, etc.) to feed data to the same service without any changes.

---

### Implementation Details

#### 1. Configuration (`internal/config/`)

**YAML-based configuration** with environment variable overrides:

```yaml
mqtt:
  broker: tcp://localhost:1883
  clientId: go-factory-listener
  topics: ["factory/#"]
  keepAlive: 60s
  reconnectAttempts: 5
  reconnectBackoff: 1s
  reconnectBackoffMax: 30s
  defaultPayloadFormat: auto
  disconnectTimeout: 250ms
```

**Environment variables** override YAML values:
- `MQTT_BROKER` - Broker URL
- `MQTT_CLIENT_ID` - Unique client identifier
- `MQTT_TOPICS` - Comma-separated topic list
- `MQTT_KEEP_ALIVE` - Keep-alive interval in seconds
- `MQTT_RECONNECT_ATTEMPTS` - Number of reconnection attempts
- `MQTT_RECONNECT_BACKOFF` - Initial backoff duration in seconds
- `MQTT_RECONNECT_BACKOFF_MAX` - Maximum backoff duration in seconds
- `MQTT_DEFAULT_PAYLOAD_FORMAT` - Payload format (auto|json|csv|raw)
- `MQTT_DISCONNECT_TIMEOUT` - Graceful disconnect timeout in milliseconds

#### 2. Multi-Format Payload Support (`internal/adapters/mqtt/payload.go`)

The subscriber automatically parses messages in multiple formats:

**JSON** (recommended):
```json
{
  "id": "sensor-01",
  "temperature": 23.5,
  "timestamp": 1710585600
}
```

**CSV**:
```
sensor-01,23.5,1710585600
```

**Raw bytes**:
```
sensor-01
```
(temperature defaults to 0, timestamp to current time)

**Auto-detection**: The parser attempts to identify the format automatically:
1. Try JSON parsing (if has "id" field)
2. Try CSV parsing (if has 2+ comma-separated fields with numeric second field)
3. Fall back to raw bytes

#### 3. Message Processing Flow

1. **Subscribe**: Subscriber connects to broker and subscribes to `factory/#` (and other configured topics)
2. **Receive**: MQTT client receives messages on subscribed topics
3. **Queue**: Messages are queued asynchronously into a buffered channel (100-message buffer)
4. **Parse**: Message processor dequeues and parses payload using configured format
5. **Process**: Parsed `SensorData` is passed to `IngestionService.ProcessReading()`
6. **Persist**: Service saves data through `SensorRepository` interface
7. **Alert**: Service triggers alerts through `AlertService` interface if temperature > 100°C

#### 4. Connection Resilience

**Exponential backoff reconnection**:
- Initial backoff: 1 second
- Each failed attempt: backoff *= 1.5 (capped at 30 seconds)
- Max attempts: 5
- Formula: `backoff = min(initial * 1.5^attempt, max)`

Example sequence:
- Attempt 1: Wait 1.0s, reconnect
- Attempt 2: Wait 1.5s, reconnect
- Attempt 3: Wait 2.25s, reconnect
- Attempt 4: Wait 3.375s, reconnect
- Attempt 5: Wait 5.06s, reconnect
- If all fail: Log failure and stop

#### 5. Graceful Shutdown

On `Ctrl+C` (SIGINT or SIGTERM):
1. Stop reconnection attempts
2. Stop message processing (drain in-flight messages)
3. Disconnect from MQTT broker (250ms timeout)
4. Exit cleanly

---

### Running the Server

#### Prerequisites

1. **MQTT Broker** (Mosquitto recommended):
   ```bash
   # macOS
   brew install mosquitto
   brew services start mosquitto

   # Docker
   docker run -d -p 1883:1883 eclipse-mosquitto
   ```

2. **Configuration file** at `./config/mqtt.yaml` (see example in repo)

#### Start Server

```bash
cd smart-factory
go run ./cmd/server
```

Expected output:
```
=== Smart Factory Server Startup ===
[CONFIG] Loading MQTT configuration from config/mqtt.yaml...
[CONFIG] MQTT Broker: tcp://localhost:1883
[CONFIG] Topics: [factory/#]
[INIT] Initializing adapters...
[INIT] ✓ Repository adapter initialized (in-memory)
[INIT] ✓ Alert adapter initialized (mock)
[INIT] Initializing core service...
[INIT] ✓ Ingestion service initialized
[INIT] Initializing MQTT subscriber...
[MQTT] Connected to broker: tcp://localhost:1883
[MQTT] Subscribed to topic: factory/#
[INIT] ✓ MQTT subscriber initialized and connected

=== Demo: Processing Test Reading ===
[DEMO] Processing reading...
[CRITICAL ALERT] Sensor turbine-X1 is at 105.50°C
[DEMO] ✓ Demo reading processed

=== Server Running ===
[INFO] Press Ctrl+C to shutdown...
```

#### Publish Test Messages

```bash
# JSON format
mosquitto_pub -t factory/sensors/sensor-01 \
  -m '{"id":"sensor-01","temperature":23.5,"timestamp":1710585600}'

# CSV format
mosquitto_pub -t factory/sensors/sensor-02 -m 'sensor-02,45.0,1710585600'

# Raw format
mosquitto_pub -t factory/sensors/sensor-03 -m 'sensor-03'

# Critical temperature alert (> 100°C)
mosquitto_pub -t factory/sensors/turbine-x1 \
  -m '{"id":"turbine-x1","temperature":105.5,"timestamp":1710585600}'
```

Then observe logs:
```
[MQTT] Processed reading from sensor-01: 23.50°C
[MQTT] Processed reading from sensor-02: 45.00°C
[MQTT] Processed reading from sensor-03: 0.00°C
[CRITICAL ALERT] Sensor turbine-x1 is at 105.50°C
```

#### Shutdown

Press `Ctrl+C`:
```
[SHUTDOWN] Received interrupt signal, shutting down gracefully...
[MQTT] Disconnecting from broker...
[MQTT] Disconnected from broker
[SHUTDOWN] MQTT subscriber disconnected
=== Server Shutdown Complete ===
```

---

### Testing the Implementation

#### Unit Tests (Payload Parsing)

```bash
go test ./internal/adapters/mqtt -v
```

#### Integration Test (Manual)

1. Start MQTT broker
2. Run server: `go run ./cmd/server`
3. In another terminal, publish messages: `mosquitto_pub ...`
4. Observe logs and verify:
   - Messages are received
   - Formats are parsed correctly
   - Service processes readings (saves to repo, alerts on critical temps)
   - Server shuts down gracefully

#### View Metrics

After processing, the subscriber tracks:
- `MessagesReceived` - Total messages received from MQTT
- `MessagesProcessed` - Successfully processed messages
- `ProcessErrors` - Messages that failed parsing or processing

---

### Future Enhancements

1. **Metrics & Monitoring**
   - Export Prometheus metrics
   - Track processing latency
   - Alert on error rates

2. **Message Acknowledgment**
   - Persist processed message IDs
   - Prevent duplicate processing on reconnects

3. **Concurrent Message Processing**
   - Current: Sequential processing (one message at a time)
   - Future: Worker pool for parallel processing (configurable concurrency)

4. **Advanced Payload Formats**
   - Protobuf, MessagePack, XML, etc.

5. **Authentication & Encryption**
   - MQTT username/password
   - TLS/SSL certificate support
   - OAuth2 integration (for future broker types)

6. **Topic-Specific Routing**
   - Different payload formats per topic
   - Topic-specific processing logic

7. **HTTP Handler Adapter**
   - Complement MQTT with HTTP endpoints
   - Reuse same service layer

---

### Architecture Benefits

✅ **Service agnostic to message source** - Add new adapters without changing core logic  
✅ **Testable** - Mock interfaces for unit testing  
✅ **Configurable** - YAML + environment variables  
✅ **Resilient** - Auto-reconnect with exponential backoff  
✅ **Flexible input formats** - JSON, CSV, raw bytes  
✅ **Graceful shutdown** - Clean resource cleanup  
✅ **Observable** - Structured logging and metrics  

---

### Files Created/Modified

**New files:**
- `internal/config/config.go` - Config loader (YAML + env vars)
- `internal/config/mqtt.go` - MQTT config struct
- `internal/adapters/mqtt/subscriber.go` - Main MQTT adapter
- `internal/adapters/mqtt/payload.go` - Multi-format parser
- `internal/adapters/mqtt/handler.go` - Handler scaffolding
- `internal/adapters/mqtt/connection.go` - Connection scaffolding
- `internal/adapters/mqtt/reconnect.go` - Reconnect scaffolding
- `config/mqtt.yaml` - Example configuration

**Modified files:**
- `cmd/server/main.go` - Added MQTT subscriber initialization & startup
- `internal/core/ports/repository.go` - Added `IngestionService` interface
- `go.mod` - Added `gopkg.in/yaml.v3` (auto-added by go mod tidy)

**Test scaffolding:**
- `internal/adapters/mqtt/subscriber_test.go` - Ready for implementation

---

### Next Steps

1. **Test the implementation**:
   - Start MQTT broker
   - Run server
   - Publish test messages
   - Verify logs and behavior

2. **Add unit tests** for payload parsing:
   - JSON parsing
   - CSV parsing
   - Auto-detection
   - Error handling

3. **Performance optimization** (if needed):
   - Increase message channel buffer
   - Add worker pool for concurrent processing
   - Profile under load

4. **Add HTTP handler adapter** (future task)
