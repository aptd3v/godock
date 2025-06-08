# godock

godock is a declarative Go library that provides a high-level, type-safe interface for interacting with Docker. It simplifies Docker operations by providing an intuitive API that focuses on what you want to achieve rather than how to achieve it.

## Features

- 🐳 **Container Management**: Create, start, stop, and manage containers with ease
- 🏗️ **Image Operations**: Pull and build Docker images
- 🌐 **Network Management**: Create and configure Docker networks
- 💾 **Volume Management**: Handle Docker volumes
- 📝 **Declarative Configuration**: Use functional options for clear and type-safe configuration
- 🔍 **Built-in Logging**: Configurable output writers for images, stats, and logs

## Installation

```bash
go get github.com/aptd3v/godock
```

## Quick Start

Here's a simple example that pulls an Nginx image and runs it with port mapping:

```go
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

```

## Examples

### Creating a Network

```go
package main

import (
	"context"
	"log"

	"github.com/aptd3v/godock/pkg/godock"
	"github.com/aptd3v/godock/pkg/godock/network"
	"github.com/aptd3v/godock/pkg/godock/networkoptions"
)

func main() {
	ctx := context.Background()
	client, err := godock.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	network := network.NewConfig("my-network")
	network.SetOptions(
		networkoptions.Driver("bridge"),
		networkoptions.Label("environment", "development"),
		networkoptions.EnableIPV6(true),
	)

	if err := client.CreateNetwork(ctx, network); err != nil {
		log.Fatal(err)
	}
}

```

### Working with Volumes

```go
package main

import (
    "context"
    "log"
    
    "github.com/aptd3v/godock/pkg/godock"
    "github.com/aptd3v/godock/pkg/godock/volume"
    "github.com/aptd3v/godock/pkg/godock/volumeoptions"
)

func main() {
    ctx := context.Background()
    client, err := godock.NewClient(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    volume := volume.NewConfig("my-data")
    volume.SetOptions(
        volumeoptions.Driver("local"),
        volumeoptions.Label("backup", "true"),
    )
    
    if err := client.CreateVolume(ctx, volume); err != nil {
        log.Fatal(err)
    }
}
```

### Building an Image

```go
package main

import (
	"context"
	"log"

	"github.com/aptd3v/godock/pkg/godock"
	"github.com/aptd3v/godock/pkg/godock/image"
)

func main() {
	ctx := context.Background()
	client, err := godock.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Create image config from Dockerfile in current directory
	img, err := image.NewImageFromSrc("./examples/basic_build_image")
	if err != nil {
		log.Fatal(err)
	}

	// Build the image
	if err := client.BuildImage(ctx, img); err != nil {
		log.Fatal(err)
	}
}

```

## Advanced Usage

### Custom Output Writers

You can customize where the output from Docker operations is written:

```go
package main

import (
    "os"
    "github.com/aptd3v/godock/pkg/godock"
)

func main() {
    client, _ := godock.NewClient(context.Background())
    
    // Write image pull/build output to a file
    f, _ := os.Create("build.log")
    client.SetImageResponeWriter(f)
    
    // Write container stats to a custom writer
    client.SetStatsResponeWriter(customWriter)
    
    // Write container logs to stdout
    client.SetLogResponseWriter(os.Stdout)
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

