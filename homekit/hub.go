package homekit

import (
	"context"

	"golang.org/x/sync/errgroup"

	"homekit-ng/homekit/tm"
)

type Broker interface {
	Run(ctx context.Context, tm *tm.TelemetryStorage) error
}

type Hub struct {
	tm      *tm.TelemetryStorage
	brokers []Broker
}

func NewHub() *Hub {
	return &Hub{
		tm: tm.NewTelemetryStorage(),
	}
}

func (m *Hub) Telemetries() *tm.TelemetryStorage {
	return m.tm
}

func (m *Hub) AddBroker(broker Broker) {
	m.brokers = append(m.brokers, broker)
}

func (m *Hub) Run(ctx context.Context) error {
	wg, ctx := errgroup.WithContext(ctx)

	for _, broker := range m.brokers {
		wg.Go(func() error {
			return broker.Run(ctx, m.tm)
		})
	}

	return wg.Wait()
}
