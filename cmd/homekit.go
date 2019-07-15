package main

import (
	"context"
	"fmt"
	"time"

	"homekit-ng/homekit/device"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	scanner := device.NewScanner("br0")
	scanInfo, err := scanner.Scan(ctx)
	fmt.Printf("%v %v\n", scanInfo, err)
	if err != nil {
		panic(err)
	}

	for id, record := range scanInfo {
		fmt.Printf("%d %v\n", id, record)
	}
}
