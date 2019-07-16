package spy

import (
	"context"
	"net"
	"time"

	"go.uber.org/zap"
)

// SYNCheck is a check that sends TCP SYN packets to the specified address
// and waits for an ACK packet.
type SYNCheck struct {
	// Addr is the target IP:port endpoint.
	Addr string
	// Interval shows how often the check will be performed.
	Interval time.Duration
	// OnActivity is called when a tracker detects any activity on the target.
	OnActivity func()
	// Internal logger, mainly for debugging purposes.
	Log *zap.SugaredLogger
}

func (m *SYNCheck) Run(ctx context.Context) error {
	timer := time.NewTicker(m.Interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			err := m.execute(ctx)
			if err != nil {
				m.Log.Warnf("failed to execute %T: %v", m, err)
				continue
			}

			m.OnActivity()
		}
	}
}

func (m *SYNCheck) execute(ctx context.Context) error {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", m.Addr)
	if err != nil {
		return err
	}

	if err := conn.Close(); err != nil {
		return err
	}

	return nil
}
