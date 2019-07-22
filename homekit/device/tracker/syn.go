package tracker

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"

	"homekit-ng/homekit/device/neighbor"
)

// SYNCheck is a check that sends TCP SYN packets to the specified address
// and waits for an ACK packet.
type SYNCheck struct {
	// MAC address.
	MAC net.HardwareAddr
	// Locator is a MAC to IP address resolver.
	Locator neighbor.Locator
	// Port is a target port.
	Port uint16
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
	addrs, err := m.Locator.Locate(m.MAC)
	if err != nil {
		return err
	}

	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", fmt.Sprintf("%s:%d", addrs[0].IP.String(), m.Port))
	if err != nil {
		return err
	}

	if err := conn.Close(); err != nil {
		return err
	}

	return nil
}
