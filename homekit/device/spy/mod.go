package spy

import (
	"context"
	"net"
	"time"
)

type Spy struct {
	IdleTimeout time.Duration
}

func (m *Spy) IsUp(mac net.HardwareAddr) bool {
	return time.Now().Sub(m.HardwareLastSeen(mac)) < m.IdleTimeout
}

func (m *Spy) HardwareLastSeen(mac net.HardwareAddr) time.Time {
	return time.Time{}
}

func (m *Spy) Register(mac net.HardwareAddr) {
}

func (m *Spy) Run(ctx context.Context) error {
	return nil
}
