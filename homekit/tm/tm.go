package tm

import (
	"strings"
	"sync"
	"time"
)

type Topic = string
type TelemetryValue = float64

type Telemetry struct {
	Topic Topic
	Value TelemetryValue
	// Timestamp shows the time when we received the telemetry.
	//
	// This value is calculated at the server side to avoid clock skewing.
	Timestamp time.Time
}

func NewTelemetry(topic Topic, value TelemetryValue) *Telemetry {
	return &Telemetry{
		Topic:     topic,
		Value:     value,
		Timestamp: time.Now(),
	}
}

type TelemetryStorage struct {
	mu          sync.RWMutex
	telemetries map[Topic]*Telemetry
}

func NewTelemetryStorage() *TelemetryStorage {
	return &TelemetryStorage{
		telemetries: map[Topic]*Telemetry{},
	}
}

func (m *TelemetryStorage) Read(topic string) []*Telemetry {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var telemetries []*Telemetry
	for key, telemetry := range m.telemetries {
		if strings.HasPrefix(key, topic) {
			telemetries = append(telemetries, telemetry)
		}
	}

	return telemetries
}

func (m *TelemetryStorage) PutMulti(telemetries []*Telemetry) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, telemetry := range telemetries {
		m.telemetries[telemetry.Topic] = telemetry
	}
}
