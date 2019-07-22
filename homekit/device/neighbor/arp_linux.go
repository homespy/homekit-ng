package neighbor

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
)

func ReadARPCache() ([]*Record, error) {
	v, err := ioutil.ReadFile("/proc/net/arp")
	if err != nil {
		return nil, err
	}

	arpCache, err := parseARPCache(string(v))
	if err != nil {
		return nil, err
	}

	var records []*Record
	for _, arpCacheLine := range arpCache {
		records = append(records, &Record{
			Dev: arpCacheLine.Device,
			MAC: arpCacheLine.HWAddr,
			IP:  arpCacheLine.IP,
		})
	}

	return records, nil
}

type ARPNeighborLocator struct{}

func (m*ARPNeighborLocator) Locate(mac net.HardwareAddr) ([]*Record, error) {
	records, err := ReadARPCache()
	if err != nil {
		return nil, err
	}

	matchedRecords := make([]*Record, 0, len(records))
	for _, record := range records {
		if bytes.Equal(record.MAC, mac) {
			matchedRecords = append(matchedRecords, record)
		}
	}

	if len(matchedRecords) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return matchedRecords, nil
}
