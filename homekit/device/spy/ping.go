package spy

import (
	"context"
	"net"
	"time"

	"github.com/tatsushid/go-fastping"
	"go.uber.org/zap"
)

type PingCheck struct {
	// Addr is the target IP endpoint.
	Addr string
	// Interval shows how often the check will be performed.
	Interval time.Duration
	// OnActivity is called when a tracker detects any activity on the target.
	OnActivity func()
	// Log is a logger.
	Log *zap.SugaredLogger
}

func (m *PingCheck) Run(ctx context.Context) error {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", m.Addr)
	if err != nil {
		return err
	}

	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		m.OnActivity()
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
