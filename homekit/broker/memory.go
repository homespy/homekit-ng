package broker

import (
	"context"

	"homekit-ng/homekit/tm"
)


type MemoryBroker struct {
	txrx chan *tm.Telemetry
}

func NewMemoryBroker() *MemoryBroker {
	return &MemoryBroker{
		txrx: make(chan *tm.Telemetry, 128),
	}
}

func (m *MemoryBroker) Add(topic string, value float64) {
	m.txrx <- tm.NewTelemetry(topic, value)
}

func (m *MemoryBroker) Run(ctx context.Context, storage *tm.TelemetryStorage) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev := <-m.txrx:
			storage.PutMulti([]*tm.Telemetry{ev})
		}
	}
}


