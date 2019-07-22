package neighbor

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProcNetARPLine(t *testing.T) {
	v := "192.168.1.43     0x1         0x2         d0:d2:b0:9c:f7:7d     *        br0"

	info, err := parseARPCacheLine(v)
	require.NoError(t, err)
	require.NotNil(t, info)

	assert.Equal(t, net.IPv4(192, 168, 1, 43), info.IP)
	assert.Equal(t, uint64(0x1), info.HWType)
	assert.Equal(t, uint64(0x2), info.Flags)
	assert.Equal(t, net.HardwareAddr{0xd0, 0xd2, 0xb0, 0x9c, 0xf7, 0x7d}, info.HWAddr)
	assert.Equal(t, "*", info.Mask)
	assert.Equal(t, "br0", info.Device)
}

func TestParseProcNetARPLineMalformed(t *testing.T) {
	v := "# 192.168.1.43 0x1 0x2 d0:d2:b0:9c:f7:7d * br0"

	info, err := parseARPCacheLine(v)
	assert.Error(t, err)
	assert.Nil(t, info)
}

func TestParseProcNetARPLineInvalidIP(t *testing.T) {
	v := "666.168.1.43 0x1 0x2 d0:d2:b0:9c:f7:7d * br0"

	info, err := parseARPCacheLine(v)
	assert.Error(t, err)
	assert.Nil(t, info)
}

func TestParseProcNetARPLineInvalidHWType(t *testing.T) {
	v := "192.168.1.43 Z 0x2 d0:d2:b0:9c:f7:7d * br0"

	info, err := parseARPCacheLine(v)
	assert.Error(t, err)
	assert.Nil(t, info)
}

func TestParseProcNetARPLineInvalidFlags(t *testing.T) {
	v := "192.168.1.43 0x1 X d0:d2:b0:9c:f7:7d * br0"

	info, err := parseARPCacheLine(v)
	assert.Error(t, err)
	assert.Nil(t, info)
}

func TestParseProcNetARPLineInvalidMAC(t *testing.T) {
	v := "192.168.1.43 0x1 0x2 xx:xx:b0:9c:f7:7d * br0"

	info, err := parseARPCacheLine(v)
	assert.Error(t, err)
	assert.Nil(t, info)
}

func TestParseProcNetARP(t *testing.T) {
	v := `
IP address       HW type     Flags       HW address            Mask     Device
192.168.1.43     0x1         0x2         d0:d2:b0:9c:f7:7d     *        br0
192.168.1.32     0x1         0x2         a4:d9:31:d0:38:e9     *        br0
192.168.1.71     0x1         0x2         08:ea:40:37:07:8a     *        br0
192.168.1.249    0x1         0x2         dc:a9:04:97:9d:9b     *        br0
192.168.1.146    0x1         0x2         54:e4:3a:93:b2:86     *        br0
128.68.64.1      0x1         0x2         a4:a1:c2:28:cf:b3     *        eth0
169.254.39.159   0x1         0x2         08:62:66:92:26:3c     *        br0
192.168.1.4      0x1         0x2         a8:66:7f:39:7b:74     *        br0`

	info, err := parseARPCache(v)
	require.NoError(t, err)
	require.NotNil(t, info)

	assert.Equal(t, 8, len(info))
}
