package broker

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"homekit-ng/homekit/tm"
)

// Decoder decodes the UDP packet into telemetry values.
//
// The format is: "<topic>=<value>";
type decoder struct{}

func (m *decoder) Decode(v string) ([]*tm.Telemetry, error) {
	var tmVec []*tm.Telemetry
	for _, kv := range strings.Split(v, ";") {
		kv = strings.TrimSpace(kv)

		if len(kv) == 0 {
			continue
		}

		parts := strings.Split(kv, "=")

		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid telemetry key-value pair")
		}

		topic := parts[0]
		value, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid telemetry value: %v", err)
		}

		tmVec = append(tmVec, tm.NewTelemetry(topic, value))
	}

	return tmVec, nil
}

type udpBroker struct {
	port uint16
	log  *zap.SugaredLogger
}

func NewUDPBroker(port uint16, log *zap.SugaredLogger) *udpBroker {
	return &udpBroker{
		port: port,
		log:  log,
	}
}

func (m *udpBroker) Run(ctx context.Context, tm *tm.TelemetryStorage) error {
	addr := fmt.Sprintf("0.0.0.0:%d", m.port)

	sock, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		return m.run(ctx, sock, tm)
	})

	<-ctx.Done()
	if err := sock.Close(); err != nil {
		m.log.Warnf("failed to close UDP socket: %v", err)
	}

	return wg.Wait()
}

// This function MUST never finish with "nil" error.
func (m *udpBroker) run(ctx context.Context, sock net.PacketConn, tm *tm.TelemetryStorage) error {
	buf := make([]byte, 4096)
	decoder := &decoder{}

	for {
		nRead, remoteAddr, err := sock.ReadFrom(buf[:])
		if err != nil {
			return err
		}

		m.log.Debugf("received %d bytes from %s", nRead, remoteAddr)

		values, err := decoder.Decode(string(buf[:nRead]))
		if err != nil {
			m.log.Warnf("failed to parse datagram: %v", err)
			continue
		}

		tm.PutMulti(values)
	}
}
