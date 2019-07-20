// +build linux

package tracker

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"go.uber.org/zap"
	"golang.org/x/net/bpf"
	"golang.org/x/sync/errgroup"
)

const (
	// Snapshot length.
	snapLen = 1024
)

type PCapCheck struct {
	// Name of captured device.
	Device string
	// BPF Filter.
	Filter []bpf.RawInstruction
	// Logger.
	Log *zap.SugaredLogger
}

func (m *PCapCheck) Run(ctx context.Context, onActivity func()) error {
	handle, err := pcapgo.NewEthernetHandle(m.Device)
	if err != nil {
		return err
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		if err := handle.SetBPF(m.Filter); err != nil {
			return err
		}

		packetSource := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
		for packet := range packetSource.Packets() {
			m.Log.Debugf("captured packet: %v", packet)
			onActivity()
		}

		return fmt.Errorf("unexpected end of packets")
	})

	<-ctx.Done()
	handle.Close()

	return wg.Wait()
}

func newBPFInstructions(mac net.HardwareAddr) []bpf.Instruction {
	return []bpf.Instruction{
		bpf.LoadAbsolute{Off: 8, Size: 4},
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: binary.BigEndian.Uint32(mac[2:6]), SkipTrue: 0, SkipFalse: 3},
		bpf.LoadAbsolute{Off: 6, Size: 2},
		bpf.JumpIf{Cond: bpf.JumpEqual, Val: uint32(binary.BigEndian.Uint16(mac[0:2])), SkipTrue: 0, SkipFalse: 1},
		bpf.RetConstant{Val: snapLen},
		bpf.RetConstant{Val: 0},
	}
}

func newBPFRawInstructions(mac net.HardwareAddr) ([]bpf.RawInstruction, error) {
	instructions := newBPFInstructions(mac)

	rawInstructions := make([]bpf.RawInstruction, len(instructions))
	for id, inst := range instructions {
		rawInstruction, err := inst.Assemble()
		if err != nil {
			return nil, err
		}

		rawInstructions[id] = rawInstruction
	}

	return rawInstructions, nil
}

func CompileFilter(mac net.HardwareAddr) ([]bpf.RawInstruction, error) {
	return newBPFRawInstructions(mac)
}
