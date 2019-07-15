package device

import (
	"context"

	"homekit-ng/homekit/device/scan"
)

type Scanner struct {
	// Interface name to scan.
	ifName string
}

// NewScanner constructs a new device scanner.
//
// We accept only a single interface name to avoid ARP storm.
// If you want to scan more than one interface it is recommended to construct
// several scanners.
func NewScanner(ifName string) *Scanner {
	return &Scanner{
		ifName: ifName,
	}
}

// Scan scans the local network for devices.
func (m *Scanner) Scan(ctx context.Context) ([]*scan.ARPCacheRecord, error) {
	return scan.ReadARPCache(ctx, m.ifName)
}
