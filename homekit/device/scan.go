package device

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

type ScanInfo struct {
	MAC string
	IP  net.IP
}

func parseScanInfoFromString(v string) (*ScanInfo, error) {
	parts := strings.Split(v, " ")
	if len(parts) < 7 {
		return nil, fmt.Errorf("malformed ARP output")
	}

	mac := parts[3]

	ipString := strings.Trim(parts[1], "()")
	ip := net.ParseIP(ipString)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address")
	}

	m := &ScanInfo{
		MAC: mac,
		IP:  ip,
	}

	return m, nil
}

func parseARPIntoScanInfo(out string) ([]*ScanInfo, error) {
	var scanInfoList []*ScanInfo
	for _, line := range strings.Split(out, "\n") {
		if len(line) == 0 {
			continue
		}

		scanInfo, err := parseScanInfoFromString(line)
		if err != nil {
			return nil, err
		}

		scanInfoList = append(scanInfoList, scanInfo)
	}

	return scanInfoList, nil
}

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
func (m *Scanner) Scan(ctx context.Context) (*ScanInfo, error) {
	cmd := exec.CommandContext(ctx, "arp", "-an", "-i", m.ifName)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseScanInfoFromString(string(out))
}
