package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
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
	// Get the absolute path for MongoDB data
	projectRoot, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}
	mongoDataDir := filepath.Join(projectRoot, "data", "mongodb")

	// Create MongoDB data directory with all parent directories
	if err := os.MkdirAll(mongoDataDir, 0755); err != nil {
		log.Fatalf("Failed to create MongoDB data directory: %v", err)
	}
	log.Printf("MongoDB data will be stored in: %s\n", mongoDataDir)

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

	// Create a network for MongoDB
	net := network.NewConfig("mongo-net")
	net.SetOptions(
		networkoptions.Driver("bridge"),
		networkoptions.Label("app", "mongodb-example"),
		networkoptions.Attachable(),
	)

	if err := client.CreateNetwork(ctx, net); err != nil {
		log.Fatalf("Failed to create network: %v", err)
	}

	// Pull MongoDB image
	image, err := image.NewConfig("mongo:latest")
	if err != nil {
		log.Fatalf("Failed to create image: %v", err)
	}
	err = client.PullImage(ctx, image)
	if err != nil {
		log.Fatalf("Failed to pull image: %v", err)
	}

	// Create MongoDB container configuration
	mongo := container.NewConfig("mongodb-server")

	// Set container options
	mongo.SetContainerOptions(
		// Set the image
		containeroptions.Image(image),

		// Expose MongoDB port
		containeroptions.Expose("27017"),

		// Add labels
		containeroptions.Label("app", "mongodb"),
		containeroptions.Label("environment", "example"),

		// Configure health check
		containeroptions.HealthCheckExec(
			time.Second*30, // start period (MongoDB needs more time to start)
			time.Second*5,  // timeout
			time.Second*10, // interval
			3,              // retries
			"CMD", "mongosh", "--eval", "db.adminCommand('ping')",
		),

		// Set working directory
		containeroptions.WorkingDir("/data/db"),
	)

	// Set host options
	mongo.SetHostOptions(
		// Configure port binding
		hostoptions.PortBindings("0.0.0.0", "27017", "27017"),

		// Set restart policy
		hostoptions.RestartAlways(),

		// Set memory limit (2GB)
		hostoptions.Memory(2*1024*1024*1024),

		// Create a bind mount for easy host access using absolute path
		hostoptions.Mount(
			hostoptions.MountType("bind"),
			mongoDataDir, // absolute host path
			"/data/db",   // container path
			false,        // read-only
		),

		// Connect to the mongo network
		hostoptions.NetworkMode("bridge"),
	)

	// Configure network endpoint settings
	endpoint := endpointoptions.NewConfig()
	endpoint.SetEndpointSetting(
		endpointoptions.Aliases("mongodb"),
	)

	// Set network options
	mongo.SetNetworkOptions(
		networkoptions.Endpoint("mongo-net", endpoint),
	)

	// Ensure cleanup on exit
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Stop and remove the container first
		if err := client.StopContainer(cleanupCtx, mongo); err != nil {
			log.Printf("Warning: Failed to stop container: %v", err)
		}
		if err := client.RemoveContainer(cleanupCtx, mongo, true); err != nil {
			log.Printf("Warning: Failed to remove container: %v", err)
		}

		// Then remove the network
		if err := client.RemoveNetwork(cleanupCtx, net); err != nil {
			log.Printf("Warning: Failed to remove network: %v", err)
		} else {
			log.Println("Successfully cleaned up network")
		}
	}()

	// Create and start MongoDB container
	if err := client.CreateContainer(ctx, mongo); err != nil {
		log.Fatalf("Failed to create MongoDB container: %v", err)
	}

	if err := client.StartContainer(ctx, mongo); err != nil {
		log.Fatalf("Failed to start MongoDB container: %v", err)
	}

	log.Printf("MongoDB container started successfully!")
	log.Printf("Connect using: mongosh mongodb://localhost:27017")
	log.Printf("Other containers in the mongo-net network can connect using: mongodb://mongodb:27017")
	log.Printf("Data directory is available at: %s", mongoDataDir)
	log.Printf("Press Ctrl+C to stop and cleanup")

	// Wait for context cancellation (interrupt signal)
	<-ctx.Done()
}
