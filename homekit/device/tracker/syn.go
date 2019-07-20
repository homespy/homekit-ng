package tracker

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
	// Internal logger, mainly for debugging purposes.
	Log *zap.SugaredLogger
}

// OnActivity is called when a tracker detects any activity on the target.
func (m *SYNCheck) Run(ctx context.Context, onActivity func()) error {
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

			onActivity()
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
