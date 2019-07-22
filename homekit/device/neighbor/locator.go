package neighbor

import (
	"net"
)

type Record struct {
	Dev string
	MAC net.HardwareAddr
	IP  net.IP
}

// Neighbor Discovery Protocol Locator.
type Locator interface {
	Locate(mac net.HardwareAddr) ([]*Record, error)
}
