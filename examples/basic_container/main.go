package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aptd3v/godock/pkg/godock"
	"github.com/aptd3v/godock/pkg/godock/container"
	"github.com/aptd3v/godock/pkg/godock/containeroptions"
	"github.com/aptd3v/godock/pkg/godock/hostoptions"
	"github.com/aptd3v/godock/pkg/godock/image"
)

func main() {
	ctx := context.Background()

	// Create a new Docker client
	client, err := godock.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	// Configure the image to pull
	img := image.NewConfig("nginx:latest")

	// Pull the image
	fmt.Println("Pulling nginx image...")
	if err := client.PullImage(ctx, img); err != nil {
		log.Fatalf("Failed to pull image: %v", err)
	}

	// Configure the container
	container := container.NewConfig("my-nginx")
	container.SetContainerOptions(
		containeroptions.Image(img),
		containeroptions.Expose("8080"),
		containeroptions.Env("NGINX_PORT", "8080"),
	)
	container.SetHostOptions(
		hostoptions.PortBindings("0.0.0.0", "8080", "8080"),
	)
	// Create the container
	fmt.Println("Creating container...")
	if err := client.CreateContainer(ctx, container); err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}

	// Start the container
	fmt.Println("Starting container...")
	if err := client.StartContainer(ctx, container); err != nil {
		log.Fatalf("Failed to start container: %v", err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		Cleanup(client, container)
		os.Exit(1)
	}()
	fmt.Println("Container is running! Press Ctrl+C to stop...")

	// Wait for a while to see the container running
	time.Sleep(10 * time.Second)
	Cleanup(client, container)
}

func Cleanup(client *godock.Client, container *container.ContainerConfig) {
	ctx := context.Background()

	// Stop the container
	fmt.Println("Stopping container...")
	if err := client.StopContainer(ctx, container); err != nil {
		log.Fatalf("Failed to stop container: %v", err)
	}

	// Remove the container
	fmt.Println("Removing container...")
	if err := client.RemoveContainer(ctx, container, true); err != nil {
		log.Fatalf("Failed to remove container: %v", err)
	}

	fmt.Println("Example completed successfully!")
}
