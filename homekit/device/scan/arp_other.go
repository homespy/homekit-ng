// +build !linux

package scan

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

func parseScanInfoFromString(v string) (*ARPCacheRecord, error) {
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

	m := &ARPCacheRecord{
		MAC: mac,
		IP:  ip,
	}

	return m, nil
}

func parseARPIntoScanInfo(out string) ([]*ARPCacheRecord, error) {
	var records []*ARPCacheRecord
	for _, line := range strings.Split(out, "\n") {
		if len(line) == 0 {
			continue
		}

		record, err := parseScanInfoFromString(line)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	return records, nil
}

func ReadARPCache(ctx context.Context, ifName string) ([]*ARPCacheRecord, error) {
	cmd := exec.CommandContext(ctx, "arp", "-an", "-i", ifName)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return parseARPIntoScanInfo(string(out))
}
