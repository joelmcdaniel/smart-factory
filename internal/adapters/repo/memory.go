package repo

import (
	"context"
	"sync"

	"github.com/joelmcdaniel/go-microservices-and-iot-injest/smart-factory/internal/core/domain"
)

// MemoryRepo implements ports.SensorRepository using a map.
// Great for testing or local development.
type MemoryRepo struct {
	mu   sync.RWMutex
	data map[string]domain.SensorData
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		data: make(map[string]domain.SensorData),
	}
}

func (m *MemoryRepo) Save(ctx context.Context, d domain.SensorData) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[d.ID] = d
	return nil
}

func (m *MemoryRepo) GetByID(ctx context.Context, id string) (*domain.SensorData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if val, ok := m.data[id]; ok {
		return &val, nil
	}
	return nil, nil // Not found
}
