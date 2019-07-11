package device

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseScanInfoOSX(t *testing.T) {
	out := "? (192.168.1.1) at fc:ec:da:41:8d:19 on en0 ifscope [ethernet]"
	info, err := parseScanInfoFromString(out)

	assert.NoError(t, err)
	require.NotNil(t, info)

	assert.Equal(t, "fc:ec:da:41:8d:19", info.MAC)
	assert.Equal(t, net.IPv4(192, 168, 1, 1), info.IP)
}

func TestParseScanInfoLinux(t *testing.T) {
	out := "? (172.17.0.2) at 02:42:ac:11:00:02 [ether] on docker0"
	info, err := parseScanInfoFromString(out)

	assert.NoError(t, err)
	require.NotNil(t, info)

	assert.Equal(t, "02:42:ac:11:00:02", info.MAC)
	assert.Equal(t, net.IPv4(172, 17, 0, 2), info.IP)
}

func TestARPIntoScanInfoOSX(t *testing.T) {
	out := `
? (192.168.1.1) at fc:ec:da:41:8d:19 on en0 ifscope [ethernet]
? (192.168.1.179) at 64:5a:ed:ea:16:2d on en0 ifscope [ethernet]
? (192.168.1.218) at a4:d9:31:d0:38:e9 on en0 ifscope [ethernet]
? (192.168.1.243) at 8:f6:9c:4f:b3:4b on en0 ifscope [ethernet]
? (224.0.0.251) at 1:0:5e:0:0:fb on en0 ifscope permanent [ethernet]
? (239.255.255.250) at 1:0:5e:7f:ff:fa on en0 ifscope permanent [ethernet]`
	info, err := parseARPIntoScanInfo(out)

	assert.NoError(t, err)
	require.NotNil(t, info)

	assert.Len(t, info, 6)
	assert.Equal(t, "fc:ec:da:41:8d:19", info[0].MAC)
	assert.Equal(t, "64:5a:ed:ea:16:2d", info[1].MAC)
	assert.Equal(t, "a4:d9:31:d0:38:e9", info[2].MAC)
	assert.Equal(t, "8:f6:9c:4f:b3:4b", info[3].MAC)
	assert.Equal(t, "1:0:5e:0:0:fb", info[4].MAC)
	assert.Equal(t, "1:0:5e:7f:ff:fa", info[5].MAC)
}
