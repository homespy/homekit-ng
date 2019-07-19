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
	"homekit-ng/homekit/device/spy"
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

func _main(path string) error {
	cfg, err := homekit.LoadConfig(path)
	if err != nil {
		return err
	}

	ctx := context.Background()
	log, err := newLogger(cfg.Logging)
	if err != nil {
		return err
	}

	log.Info("initialized HomeKit", zap.Any("config", cfg))

	// todo: from cfg.
	dev := "br0"
	mac, err := net.ParseMAC("a4:d9:31:d0:38:e9")
	if err != nil {
		return err
	}

	s := spy.NewSpy(dev, log.Sugar())
	s.Register(mac)

	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(func() error {
		timer := time.NewTicker(5 * time.Second)
		defer timer.Stop()

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				log.Sugar().Infof("%s last seen: %s ago", mac, time.Now().Sub(s.HardwareLastSeen(mac)))
			}
		}
	})
	wg.Go(func() error {
		return s.Run(ctx)
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
		return _main(cmd.String("config"))
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
}
