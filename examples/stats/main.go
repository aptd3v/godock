package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aptd3v/godock/pkg/godock"
	"github.com/aptd3v/godock/pkg/godock/container"
	"github.com/aptd3v/godock/pkg/godock/containeroptions"
	"github.com/aptd3v/godock/pkg/godock/image"
)

func main() {
	ctx := context.Background()
	fmt.Println("Starting stats example...")
	// Create a new Docker client
	client, err := godock.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	// Configure and pull Alpine image
	alpineImg := image.NewConfig("alpine")

	rc, err := client.ImagePull(ctx, alpineImg)
	if err != nil {
		log.Fatalf("Failed to pull image: %v", err)
	}
	defer rc.Close()
	_, err = io.Copy(os.Stdout, rc)
	if err != nil {
		log.Fatalf("failed to copy logs: %v", err)
	}

	// Configure the container to do some work
	container := container.NewConfig("stats-test")
	container.SetContainerOptions(
		containeroptions.Image(alpineImg),
		// Generate some CPU and network activity
		containeroptions.CMD("sh", "-c", "while true; do wget -q -O- https://example.com > /dev/null; sleep 1; done"),
	)

	// Create and start the container
	fmt.Println("Creating and starting container...")
	if err := client.ContainerCreate(ctx, container); err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}
	if err := client.ContainerStart(ctx, container); err != nil {
		log.Fatalf("Failed to start container: %v", err)
	}

	// Setup cleanup on interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup(client, container)
		os.Exit(1)
	}()

	// Get stats stream
	statsCh, errCh := client.ContainerStatsChan(ctx, container)

	// Print stats every second
	fmt.Println("Monitoring container stats (Ctrl+C to exit)...")
	for {
		select {
		case stats, ok := <-statsCh:
			if !ok {
				return
			}
			fmt.Printf("\nContainer Stats at %s:\n", time.Now().Format(time.RFC3339))
			fmt.Printf("CPU Usage: %s\n", stats.FormatCpuUsagePercentage())
			fmt.Printf("Memory Usage: %s\n", stats.FormatMemoryUsage())
			fmt.Printf("Network I/O: %s\n", stats.FormatNetworkIO())
		case err := <-errCh:
			if err != nil {
				log.Printf("Stats error: %v\n", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func cleanup(client *godock.Client, container *container.ContainerConfig) {
	ctx := context.Background()
	fmt.Println("\nCleaning up...")
	if err := client.ContainerRemove(ctx, container, true); err != nil {
		log.Printf("Failed to remove container: %v", err)
	}
}
