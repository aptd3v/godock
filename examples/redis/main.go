package main

import (
	"context"
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
	"github.com/aptd3v/godock/pkg/godock/network"
	"github.com/aptd3v/godock/pkg/godock/networkoptions"
	"github.com/aptd3v/godock/pkg/godock/networkoptions/endpointoptions"
)

func main() {
	// Create a context that we can cancel on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received interrupt signal, cleaning up...")
		cancel()
	}()

	// Initialize Docker client
	client, err := godock.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	// Create a network for Redis
	net := network.NewConfig("redis-net")
	net.SetOptions(
		networkoptions.Driver("bridge"),
		networkoptions.Label("app", "redis-example"),
		networkoptions.Attachable(),
	)

	if err := client.CreateNetwork(ctx, net); err != nil {
		log.Fatalf("Failed to create network: %v", err)
	}

	// Pull Redis image
	image, err := image.NewConfig("redis")
	if err != nil {
		log.Fatalf("Failed to create image: %v", err)
	}
	err = client.PullImage(ctx, image)
	if err != nil {
		log.Fatalf("Failed to pull image: %v", err)
	}

	// Create Redis container configuration
	redis := container.NewConfig("redis-server")

	// Set container options
	redis.SetContainerOptions(
		// Set the image
		containeroptions.Image(image),

		// Expose Redis port
		containeroptions.Expose("6379"),

		// Add labels
		containeroptions.Label("app", "redis"),
		containeroptions.Label("environment", "example"),

		// Configure health check
		containeroptions.HealthCheckExec(
			time.Second*5, // start period
			time.Second*3, // timeout
			time.Second*5, // interval
			3,             // retries
			"CMD", "redis-cli", "ping",
		),

		// Set working directory
		containeroptions.WorkingDir("/data"),
	)

	// Set host options
	redis.SetHostOptions(
		// Configure port binding
		hostoptions.PortBindings("0.0.0.0", "6379", "6379"),

		// Set restart policy
		hostoptions.RestartAlways(),

		// Set memory limit (1GB)
		hostoptions.Memory(1*1024*1024*1024),

		// Mount volume for persistence
		hostoptions.Mount(hostoptions.MountType("volume"), "redis-data", "/data", false),

		// Connect to the redis network
		hostoptions.NetworkMode("bridge"),
	)

	// Configure network endpoint settings
	endpoint := endpointoptions.NewConfig()
	endpoint.SetEndpointSetting(
		endpointoptions.Aliases("redis"),
	)

	// Set network options
	redis.SetNetworkOptions(
		networkoptions.Endpoint("redis-net", endpoint),
	)

	// Ensure cleanup on exit
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Stop and remove the container first
		if err := client.StopContainer(cleanupCtx, redis); err != nil {
			log.Printf("Warning: Failed to stop container: %v", err)
		}
		if err := client.RemoveContainer(cleanupCtx, redis, true); err != nil {
			log.Printf("Warning: Failed to remove container: %v", err)
		}

		// Then remove the network
		if err := client.RemoveNetwork(cleanupCtx, net); err != nil {
			log.Printf("Warning: Failed to remove network: %v", err)
		} else {
			log.Println("Successfully cleaned up network")
		}
	}()

	// Create and start Redis container
	if err := client.CreateContainer(ctx, redis); err != nil {
		log.Fatalf("Failed to create Redis container: %v", err)
	}

	if err := client.StartContainer(ctx, redis); err != nil {
		log.Fatalf("Failed to start Redis container: %v", err)
	}

	log.Printf("Redis container started successfully!")
	log.Printf("Connect using: redis-cli -h localhost -p 6379")
	log.Printf("Other containers in the redis-net network can connect using: redis-cli -h redis")
	log.Printf("Press Ctrl+C to stop and cleanup")

	// Wait for context cancellation (interrupt signal)
	<-ctx.Done()
}
