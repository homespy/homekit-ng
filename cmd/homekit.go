package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"homekit-ng/homekit/device"
	"homekit-ng/homekit/device/spy"
)

func _main() error {
	ctx := context.Background()

	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	log, err := cfg.Build()
	if err != nil {
		return err
	}

	// todo: from cfg.
	dev := "en0"
	mac := "a4:d9:31:d0:38:e9"

	scanner := device.NewScanner(dev)
	scanInfo, err := scanner.Scan(ctx)
	if err != nil {
		return err
	}

	ip := net.IP{}
	for id, record := range scanInfo {
		fmt.Printf("%d %s -> %s\n", id, record.MAC, record.IP.String())
		if mac == record.MAC {
			ip = record.IP
		}
	}

	wg, ctx := errgroup.WithContext(ctx)

	//hub := homekit.NewHub()
	//hub.AddBroker(broker.NewUDPBroker(9090, log.Sugar()))
	//
	//wg.Go(func() error {
	//	for {
	//		log.Debug("reading '/home")
	//		telemetries := hub.Telemetries().Read("/home")
	//		for _, v := range telemetries {
	//			fmt.Printf("%s : %s=%v\n", v.Timestamp, v.Topic, v.Value)
	//		}
	//
	//		select {
	//		case <-ctx.Done():
	//			return ctx.Err()
	//		default:
	//			time.Sleep(time.Second)
	//		}
	//	}
	//})

	lastSeen := time.Unix(0, 0)

	wg.Go(func() error {
		timer := time.NewTicker(5 * time.Second)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				log.Sugar().Infof("last seen: %s ago", time.Now().Sub(lastSeen))

			}
		}
	})
	wg.Go(func() error {
		sc := &spy.PingCheck{
			Addr:     ip.String(),
			Interval: 1 * time.Second,
			OnActivity: func() {
				log.Info("detected PING activity")
				lastSeen = time.Now()
			},
			Log: log.Sugar(),
		}

		return sc.Run(ctx)
	})
	wg.Go(func() error {
		sc := &spy.SYNCheck{
			Addr:     fmt.Sprintf("%s:62078", ip.String()),
			Interval: 1 * time.Second,
			OnActivity: func() {
				log.Info("detected SYN activity")
				lastSeen = time.Now()
			},
			Log: log.Sugar(),
		}

		return sc.Run(ctx)
	})

	wg.Go(func() error {
		MAC, err := net.ParseMAC(mac)
		if err != nil {
			return err
		}

		filter, err := spy.CompileFilter(MAC)
		if err != nil {
			return err
		}

		sc := &spy.PCapCheck{
			Device: dev,
			Filter: filter,
			OnActivity: func() {
				log.Info("detected PCAP activity")
				lastSeen = time.Now()
			},
			Log: log.Sugar(),
		}

		return sc.Run(ctx)
	})

	//if err := hub.Run(ctx); err != nil {
	//	return err
	//}

	return wg.Wait()
}

func main() {
	if err := _main(); err != nil {
		panic(err)
	}
}
