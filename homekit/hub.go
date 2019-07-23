package homekit

import (
	"context"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"homekit-ng/homekit/tm"
)

type Broker interface {
	Run(ctx context.Context, tm *tm.TelemetryStorage) error
}

type Hub struct {
	tm      *tm.TelemetryStorage
	brokers []Broker
	log     *zap.SugaredLogger
}

func NewHub(log *zap.SugaredLogger) *Hub {
	return &Hub{
		tm:  tm.NewTelemetryStorage(),
		log: log,
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
		broker := broker

		wg.Go(func() error {
			m.log.Infof("running %T", broker)
			defer m.log.Infof("stopped %T", broker)

			return broker.Run(ctx, m.tm)
		})
	}

	return wg.Wait()
}
