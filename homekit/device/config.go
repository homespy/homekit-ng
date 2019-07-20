package device

import (
	"context"
	"fmt"
	"net"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"

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
		panic("unimplemented")
		//type config struct {
		//	Port     uint16
		//	Interval time.Duration
		//}
		//
		//args := &config{}
		//
		//if err := mapstructure.Decode(cfg.Args, &args); err != nil {
		//	return nil, err
		//}
		//
		//return func(cfg *networkConfig, log *zap.SugaredLogger) (Tracker, error) {
		//	return &tracker.SYNCheck{
		//		Addr:     fmt.Sprintf("%s:%d", cfg.IPAddr, args.Port),
		//		Interval: args.Interval,
		//		Log:      log,
		//	}, nil
		//}, nil
	case "ping":
		panic("unimplemented")
		//type config struct {
		//	IPAddr   string
		//	Interval time.Duration
		//}
		//
		//args := &config{}
		//
		//if err := mapstructure.Decode(cfg.Args, &args); err != nil {
		//	return nil, err
		//}
		//
		//return func(cfg *networkConfig, log *zap.SugaredLogger) (Tracker, error) {
		//	return &tracker.PingCheck{
		//		IPAddr:   args.IPAddr,
		//		Interval: args.Interval,
		//		Log:      log,
		//	}, nil
		//}, nil
	case "pcap":
		type config struct {
			Dev string
		}

		args := &config{}

		if err := mapstructure.Decode(cfg.Args, &args); err != nil {
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
