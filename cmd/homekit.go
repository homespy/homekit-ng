package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/urfave/cli"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"homekit-ng/homekit"
	"homekit-ng/homekit/broker"
	"homekit-ng/homekit/device"
)

const (
	AppName = "homekit"
)

var (
	AppVersion string
)

func newLogger(cfg homekit.LoggingConfig) (*zap.Logger, error) {
	loggingConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(cfg.Level),
		Development:      false,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	return loggingConfig.Build()
}

func run(path string) error {
	cfg, err := homekit.LoadConfig(path)
	if err != nil {
		return err
	}

	ctx := context.Background()
	log, err := newLogger(cfg.Logging)
	if err != nil {
		return err
	}

	log.Info("initialized HomeKit")

	deviceTracker := device.NewActivityTracker(log.Sugar())
	for mac, config := range cfg.Tracking.Devices {
		mac, err := net.ParseMAC(mac)
		if err != nil {
			return err
		}

		if err := deviceTracker.Register(mac, config); err != nil {
			return err
		}
	}

	memoryBroker := broker.NewMemoryBroker()

	hub := homekit.NewHub(log.Sugar())
	hub.AddBroker(memoryBroker)
	hub.AddBroker(broker.NewUDPBroker(cfg.Broker.Port, log.Sugar()))

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		timer := time.NewTicker(5 * time.Second)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				for mac := range cfg.Tracking.Devices {
					isUp := 0.0
					if deviceTracker.IsUp(mac) {
						isUp = 1.0
					}
					memoryBroker.Add(fmt.Sprintf("/home/%s", mac), isUp)
				}
			}
		}
	})
	wg.Go(func() error {
		return deviceTracker.Run(ctx)
	})
	wg.Go(func() error {
		timer := time.NewTicker(10 * time.Second)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
			return ctx.Err()
			case <-timer.C:
				for id, telemetry := range hub.Telemetries().Read("/") {
					log.Sugar().Infof("[%d] %s: %.2f", id, telemetry.Topic, telemetry.Value)
				}
			}
		}

	})
	wg.Go(func() error {
		return hub.Run(ctx)
	})

	return wg.Wait()
}

func main() {
	app := cli.NewApp()
	app.Name = AppName
	app.Usage = "HomeKit device presence tracker and telemetry broker"
	app.Version = AppVersion
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Usage: "Load configuration from `FILE`",
			Value: "/etc/homekit/homekit.json",
		},
	}
	app.Action = func(cmd *cli.Context) error {
		return run(cmd.String("config"))
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}
