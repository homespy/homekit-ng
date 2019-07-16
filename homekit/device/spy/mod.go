package spy

import (
	"context"
	"net"
)

type Spy struct {}

func (m *Spy) Watch(ip net.IP, mac net.HardwareAddr, onActivity func()) uint64 {
	return 0
}

func (m *Spy) Unwatch(id uint64) {}

func (m *Spy) Run(ctx context.Context) error {
	return nil
}
