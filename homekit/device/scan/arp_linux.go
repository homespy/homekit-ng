package scan

import (
	"context"
	"io/ioutil"
)

func ReadARPCache(ctx context.Context, ifName string) ([]*ARPCacheRecord, error) {
	v, err := ioutil.ReadFile("/proc/net/arp")
	if err != nil {
		return nil, err
	}

	arpCache, err := parseARPCache(string(v))
	if err != nil {
		return nil, err
	}

	var records []*ARPCacheRecord
	for _, arpCacheLine := range arpCache {
		if arpCacheLine.Device != ifName {
			continue
		}

		records = append(records, &ARPCacheRecord{
			MAC: arpCacheLine.HWAddr.String(),
			IP:  arpCacheLine.IP,
		})
	}

	return records, nil
}
