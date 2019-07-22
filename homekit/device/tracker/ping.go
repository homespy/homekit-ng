package tracker

import (
	"context"
	"net"
	"time"

	"github.com/tatsushid/go-fastping"
	"go.uber.org/zap"

	"homekit-ng/homekit/device/neighbor"
)

type PingCheck struct {
	// MAC address.
	MAC net.HardwareAddr
	// Locator is a MAC to IP address resolver.
	Locator neighbor.Locator
	// Interval shows how often the check will be performed.
	Interval time.Duration
	// Log is a logger.
	Log *zap.SugaredLogger
}

func (m *PingCheck) Run(ctx context.Context, onActivity func()) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := m.run(ctx, onActivity); err != nil {
				m.Log.Warnf("failed to execute %T: %v", m, err)
				time.Sleep(m.Interval)
				continue
			}
		}
	}
}

func (m *PingCheck) run(ctx context.Context, onActivity func()) error {
	addrs, err := m.Locator.Locate(m.MAC)
	if err != nil {
		return err
	}

	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", addrs[0].IP.String())
	if err != nil {
		return err
	}

	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		onActivity()
	}

	p.RunLoop()
	defer p.Stop()

	timer := time.NewTicker(m.Interval)
	defer timer.Stop()

	for {
		select {
		case <-p.Done():
			if err := p.Err(); err != nil {
				m.Log.Debugf("stopped %T", m)
				return err
			}
		case <-timer.C:
			continue
		}
	}
}
