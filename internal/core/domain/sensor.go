package domain

import "errors"

// SensorData is our pure business entity.
// It doesn't care about JSON tags or DB columns yet.

type SensorData struct {
	ID          string
	Temperature float64
	Timestamp   int64
}

// Business Rule: A sensor is "Critical" if temp > 100.
func (s *SensorData) IsCritical() bool {
	return s.Temperature > 100.00
}

// Validation logic
func (s *SensorData) Validate() error {
	if s.ID == "" {
		return errors.New("sensor ID is required")
	}
	return nil
}
