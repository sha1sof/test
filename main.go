package main

import (
	"context"
	"cpu/os"
	"fmt"
	"runtime"
	"time"
)

func main() {
	// For the load.
	go func() {
		i := 0
		for {
			i++
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	switch runtime.GOOS {
	case "windows":
		usage, currentProcessUsage, err := os.GetCpuUsage(ctx)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Total CPU usage: %d%%\n", usage)
		fmt.Printf("Current process CPU usage: %d%%\n", currentProcessUsage)
		time.Sleep(1 * time.Minute)
	case "linux":
		usage, currentProcessUsage, err := os.GetCpuUsage(ctx)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Total CPU usage: %d%%\n", usage)
		fmt.Printf("Current process CPU usage: %d%%\n", currentProcessUsage)
		time.Sleep(1 * time.Minute)
	}

	/*	usage, currentProcessUsage, err := os.GetCpuUsage(ctx)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Printf("Total CPU usage: %d%%\n", usage)
		fmt.Printf("Current process CPU usage: %d%%\n", currentProcessUsage)
		time.Sleep(1 * time.Minute)*/
}
