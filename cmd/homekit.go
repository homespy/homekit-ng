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

	scanner := device.NewScanner("en0")
	scanInfo, err := scanner.Scan(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", scanInfo)
}
