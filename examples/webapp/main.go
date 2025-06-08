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

type containerConfig struct {
	container *container.ContainerConfig
	image     *image.ImageConfig
}

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

	// Get project root for data directories
	projectRoot, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	// Initialize Docker client
	client, err := godock.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	// Create application network
	net := network.NewConfig("webapp-net")
	net.SetOptions(
		networkoptions.Driver("bridge"),
		networkoptions.Label("app", "webapp-example"),
		networkoptions.Attachable(),
	)

	if err := client.CreateNetwork(ctx, net); err != nil {
		log.Fatalf("Failed to create network: %v", err)
	}

	// Create data directories
	dirs := []string{
		filepath.Join(projectRoot, "data", "mongodb"),
		filepath.Join(projectRoot, "data", "redis"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create data directory %s: %v", dir, err)
		}
	}

	// Initialize containers
	containers := make(map[string]*containerConfig)

	// MongoDB Configuration
	mongoImage, err := image.NewConfig("mongo:latest")
	if err != nil {
		log.Fatalf("Failed to create MongoDB image config: %v", err)
	}
	mongoDB := container.NewConfig("mongodb")
	mongoDB.SetContainerOptions(
		containeroptions.Image(mongoImage),
		containeroptions.Expose("27017"),
		containeroptions.Label("service", "database"),
		containeroptions.HealthCheckExec(
			time.Second*30,
			time.Second*5,
			time.Second*10,
			3,
			"CMD", "mongosh", "--eval", "db.adminCommand('ping')",
		),
	)
	mongoDB.SetHostOptions(
		hostoptions.PortBindings("127.0.0.1", "27017", "27017"),
		hostoptions.RestartAlways(),
		hostoptions.Memory(1*1024*1024*1024),
		hostoptions.Mount(
			hostoptions.MountType("bind"),
			filepath.Join(projectRoot, "data", "mongodb"),
			"/data/db",
			false,
		),
		hostoptions.NetworkMode("bridge"),
	)
	containers["mongodb"] = &containerConfig{container: mongoDB, image: mongoImage}

	// Redis Configuration
	redisImage, err := image.NewConfig("redis:latest")
	if err != nil {
		log.Fatalf("Failed to create Redis image config: %v", err)
	}
	redis := container.NewConfig("redis")
	redis.SetContainerOptions(
		containeroptions.Image(redisImage),
		containeroptions.Expose("6379"),
		containeroptions.Label("service", "cache"),
		containeroptions.HealthCheckExec(
			time.Second*5,
			time.Second*3,
			time.Second*5,
			3,
			"CMD", "redis-cli", "ping",
		),
	)
	redis.SetHostOptions(
		hostoptions.PortBindings("127.0.0.1", "6379", "6379"),
		hostoptions.RestartAlways(),
		hostoptions.Memory(512*1024*1024),
		hostoptions.Mount(
			hostoptions.MountType("bind"),
			filepath.Join(projectRoot, "data", "redis"),
			"/data",
			false,
		),
		hostoptions.NetworkMode("bridge"),
	)
	containers["redis"] = &containerConfig{container: redis, image: redisImage}

	// Nginx Frontend Configuration
	nginxImage, err := image.NewConfig("nginx:latest")
	if err != nil {
		log.Fatalf("Failed to create Nginx image config: %v", err)
	}
	nginx := container.NewConfig("frontend")
	nginx.SetContainerOptions(
		containeroptions.Image(nginxImage),
		containeroptions.Expose("80"),
		containeroptions.Label("service", "frontend"),
		containeroptions.HealthCheckExec(
			time.Second*5,
			time.Second*3,
			time.Second*5,
			3,
			"CMD", "curl", "--fail", "http://localhost:80/health",
		),
	)
	nginx.SetHostOptions(
		hostoptions.PortBindings("0.0.0.0", "80", "80"),
		hostoptions.RestartAlways(),
		hostoptions.Memory(256*1024*1024),
		hostoptions.NetworkMode("bridge"),
	)
	containers["nginx"] = &containerConfig{container: nginx, image: nginxImage}

	// Configure network endpoints for service discovery
	for name, cfg := range containers {
		endpoint := endpointoptions.NewConfig()
		endpoint.SetEndpointSetting(
			endpointoptions.Aliases(name),
		)
		cfg.container.SetNetworkOptions(
			networkoptions.Endpoint("webapp-net", endpoint),
		)
	}

	// Pull images and create containers
	for name, cfg := range containers {
		log.Printf("Setting up %s...", name)

		// Pull image
		if err := client.PullImage(ctx, cfg.image); err != nil {
			log.Fatalf("Failed to pull %s image: %v", name, err)
		}

		// Create container
		if err := client.CreateContainer(ctx, cfg.container); err != nil {
			log.Fatalf("Failed to create %s container: %v", name, err)
		}

		// Start container
		if err := client.StartContainer(ctx, cfg.container); err != nil {
			log.Fatalf("Failed to start %s container: %v", name, err)
		}

		log.Printf("%s is ready!", name)
	}

	// Ensure cleanup on exit
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// Stop and remove containers
		for name, cfg := range containers {
			if err := client.StopContainer(cleanupCtx, cfg.container); err != nil {
				log.Printf("Warning: Failed to stop %s container: %v", name, err)
			}
			if err := client.RemoveContainer(cleanupCtx, cfg.container, true); err != nil {
				log.Printf("Warning: Failed to remove %s container: %v", name, err)
			}
		}

		// Remove network
		if err := client.RemoveNetwork(cleanupCtx, net); err != nil {
			log.Printf("Warning: Failed to remove network: %v", err)
		} else {
			log.Println("Successfully cleaned up network")
		}
	}()

	log.Println("\nApplication stack is ready!")
	log.Println("Services:")
	log.Println("- Frontend: http://localhost:80")
	log.Println("- MongoDB: mongodb://localhost:27017")
	log.Println("- Redis: redis://localhost:6379")
	log.Println("\nContainer Network Aliases:")
	log.Println("- Frontend: nginx")
	log.Println("- MongoDB: mongodb")
	log.Println("- Redis: redis")
	log.Println("\nPress Ctrl+C to stop and cleanup")

	// Wait for context cancellation (interrupt signal)
	<-ctx.Done()
}
