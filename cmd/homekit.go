package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"homekit-ng/homekit"
	"homekit-ng/homekit/broker"
	"homekit-ng/homekit/device"
)

func _main() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	scanner := device.NewScanner("en0")
	scanInfo, err := scanner.Scan(ctx)
	if err != nil {
		return err
	}

	for id, record := range scanInfo {
		fmt.Printf("%d %s -> %s\n", id, record.MAC, record.IP.String())
	}

	hub := homekit.NewHub()
	hub.AddBroker(broker.NewUDPBroker(9090, log.Sugar()))

	go func() {
		for {
			telemetries := hub.Telemetries().Read("/")
			for _, v := range telemetries {
				fmt.Printf("%s : %s=%v\n", v.Timestamp, v.Topic, v.Value)
			}
			time.Sleep(time.Second)
		}
	}()

	if err := hub.Run(ctx); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := _main(); err != nil {
		panic(err)
	}
}
