package scan

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type ARPCacheRecord struct {
	MAC string
	IP  net.IP
}

type ARPCacheLine struct {
	IP     net.IP
	HWType uint64
	Flags  uint64
	HWAddr net.HardwareAddr
	Mask   string
	Device string
}

func parseARPCache(v string) ([]*ARPCacheLine, error) {
	var records []*ARPCacheLine

	for id, line := range strings.Split(v, "\n") {
		if len(line) == 0 {
			continue
		}

		if strings.HasPrefix(line, "IP address") {
			// Skip header.
			continue
		}

		record, err := parseARPCacheLine(line)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %d ARP cache line: %v", id, err)
		}

		records = append(records, record)
	}

	return records, nil
}

func parseARPCacheLine(v string) (*ARPCacheLine, error) {
	parts := strings.Fields(v)
	if len(parts) != 6 {
		return nil, fmt.Errorf("malformed ARP cache line string")
	}

	ip := net.ParseIP(parts[0])
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address")
	}

	hwType, err := strconv.ParseUint(parts[1], 0, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse hardware type: %v", err)
	}

	flags, err := strconv.ParseUint(parts[2], 0, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse flags: %v", err)
	}

	hwAddr, err := net.ParseMAC(parts[3])
	if err != nil {
		return nil, fmt.Errorf("failed to parse MAC address: %v", err)
	}

	mask := parts[4]
	device := parts[5]

	m := &ARPCacheLine{
		IP:     ip,
		HWType: hwType,
		Flags:  flags,
		HWAddr: hwAddr,
		Mask:   mask,
		Device: device,
	}

	return m, nil
}
