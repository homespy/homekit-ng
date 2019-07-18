package main

import (
	"context"
	"net"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

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
	mac, err := net.ParseMAC("a4:d9:31:d0:38:e9")
	if err != nil {
		return err
	}

	s := spy.NewSpy(dev, log.Sugar())
	s.Register(mac)

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


	wg.Go(func() error {
		timer := time.NewTicker(5 * time.Second)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				log.Sugar().Infof("last seen: %s ago", time.Now().Sub(s.HardwareLastSeen(mac)))
			}
		}
	})
	wg.Go(func() error {
		return s.Run(ctx)
	})

	return wg.Wait()
}

func main() {
	if err := _main(); err != nil {
		panic(err)
	}
}
