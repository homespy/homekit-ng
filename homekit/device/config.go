package device

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"homekit-ng/homekit/device/neighbor"
	"homekit-ng/homekit/device/tracker"
)

type TrackingConfig struct {
	Methods []*TrackingMethodConfig
}

type TrackingMethodConfig struct {
	Type string      `json:"type"`
	Args interface{} `json:"args"`
}

type Tracker interface {
	Run(ctx context.Context, onActivity func()) error
}

type TrackingMethodFactory = func(mac net.HardwareAddr, log *zap.SugaredLogger) (Tracker, error)

func NewTrackingMethodFactory(cfg *TrackingMethodConfig) (TrackingMethodFactory, error) {
	switch cfg.Type {
	case "syn":
		type config struct {
			Dev      string
			Port     uint16
			Interval time.Duration
		}

		args := &config{}
		if err := transcode(cfg.Args, &args); err != nil {
			return nil, err
		}

		return func(mac net.HardwareAddr, log *zap.SugaredLogger) (Tracker, error) {
			return &tracker.SYNCheck{
				MAC:      mac,
				Locator:  &neighbor.ARPNeighborLocator{},
				Port:     args.Port,
				Interval: args.Interval,
				Log:      log,
			}, nil
		}, nil
	case "ping":
		type config struct {
			Dev      string
			Interval time.Duration
		}

		args := &config{}
		if err := transcode(cfg.Args, &args); err != nil {
			return nil, err
		}

		return func(mac net.HardwareAddr, log *zap.SugaredLogger) (Tracker, error) {
			return &tracker.PingCheck{
				MAC:      mac,
				Locator:  &neighbor.ARPNeighborLocator{},
				Interval: args.Interval,
				Log:      log,
			}, nil
		}, nil
	case "pcap":
		type config struct {
			Dev string
		}

		args := &config{}
		if err := transcode(cfg.Args, &args); err != nil {
			return nil, err
		}

		return func(mac net.HardwareAddr, log *zap.SugaredLogger) (Tracker, error) {
			filter, err := tracker.CompileFilter(mac)
			if err != nil {
				return nil, err
			}

			return &tracker.PCapCheck{
				Device: args.Dev,
				Filter: filter,
				Log:    log,
			}, nil
		}, nil
	default:
		return nil, fmt.Errorf("unknown tracking method: %s", cfg.Type)
	}
}

func transcode(v, o interface{}) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, o)
}
