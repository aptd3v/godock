# godock

A type-safe, developer-friendly wrapper around Docker's Go SDK.

## Why godock?

godock was created to make Docker container management in Go more intuitive, type-safe, and maintainable. Here's why you might want to use it:

### 🛡️ Type Safety First
```go
// godock enforces type safety through dedicated option functions
redis.SetContainerOptions(
	containeroptions.Image(redisImage),      // Type-safe image configuration
	containeroptions.Expose("6379"),         // Port validation
	containeroptions.Label("env", "prod"),   // Structured labels
)

// vs Docker SDK's map-based approach
container := &container.Config{
	Image: "redis",                  // String that could be invalid
	ExposedPorts: map[string]struct{}{
		"6379/tcp": {},             // Easy to make typos
	},
	Labels: map[string]string{...}, // No structure enforcement
}
```

### 🎨 Clean, Fluent API
- Builder pattern for intuitive configuration
- Clear separation between container, host, and network options
- Self-documenting function names
- IDE-friendly with great autocompletion

### 🏗️ Structured Configuration
```go
// Clear separation of concerns
container.SetContainerOptions(...)  // Container-specific settings
container.SetHostOptions(...)       // Host-specific settings
container.SetNetworkOptions(...)    // Network-specific settings
```

### 🔒 Error Prevention
- Catch configuration errors at compile time
- Validate settings before they reach Docker
- Clear error messages for debugging
- Prevent common Docker configuration mistakes

### 📦 Resource Management
```go
// Clear, readable resource specifications
hostoptions.Memory(1*1024*1024*1024)  // Clearly 1GB
hostoptions.CPUQuota(2)               // 2 CPU cores
hostoptions.RestartAlways()           // Clear policy names
```

### 🌐 Network Simplification
```go
// Easy network configuration and service discovery
container.SetNetworkOptions(
	networkoptions.Endpoint("my-net", endpoint),
	networkoptions.Aliases("service-name"),
)
```

### 💡 Developer Experience
- Reduce boilerplate code
- Make container configuration maintainable
- Prevent common mistakes
- Great for applications managing containers programmatically

### 📚 Examples 
Check out our examples directory for usage:
- Single container setups
- Multi-container applications
- Network configuration
- Volume management
- Health checks
- And More

## Features

- 🐳 **Container Management**: Create, start, stop, and manage containers with ease
- 🏗️ **Image Operations**: Pull and build Docker images
- 🌐 **Network Management**: Create and configure Docker networks
- 💾 **Volume Management**: Handle Docker volumes
- 📝 **Declarative Configuration**: Use functional options for clear and type-safe configuration
- 🔍 **Built-in Logging**: Configurable output writers for images, stats, and logs

## Installation

To use godock in your project, you can install it using one of the following methods:

### Using go get
```bash
go get github.com/aptd3v/godock@v1.0.0
```

### Using go.mod
Add the following to your go.mod file:
```
require github.com/aptd3v/godock v1.0.0
```
Then run:
```bash
go mod tidy
```

### Requirements
- Go 1.23.0 or later
- Docker Engine running on your system

## Development

### Project Structure
```
godock/
├── examples/              # Example applications and use cases
│   ├── basic_container/   # Simple container management
│   ├── container_api/     # REST API for container management
│   └── webapp/           # Multi-container web application
├── pkg/
│   └── godock/           # Main package
│       ├── container/    # Container management
│       ├── image/        # Image operations
│       ├── network/      # Network management
│       ├── volume/       # Volume management
│       └── *options/     # Type-safe option packages
├── CONTRIBUTING.md       # Contribution guidelines
├── LICENSE              # MIT License
└── README.md           # This file
```

### Development Setup

1. **Prerequisites**
   - Go 1.21 or later
   - Docker Engine
   - Git

2. **Clone and Setup**
   ```bash
   # Clone the repository
   git clone https://github.com/YOUR-USERNAME/godock.git
   cd godock

   # Install dependencies
   go mod download
   ```

3. **Verify Setup**
   ```bash
   # Run tests
   go test ./...
   ```

### Testing

godock uses a comprehensive testing suite including both unit and integration tests:

1. **Run Unit Tests**
   ```bash
   go test ./...
   ```

2. **Run Integration Tests** (requires Docker)
   ```bash
   go test ./... -tags=integration
   ```

3. **Test Coverage**
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   ```

4. **Run Specific Tests**
   ```bash
   # Test a specific package
   go test ./pkg/godock/container/...

   # Run tests with verbose output
   go test -v ./...
   ```

### Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details on:
- How to submit changes
- Coding standards
- Testing requirements
- Pull request process

## Contributors

Thank you to all our contributors who help make godock better! 

<a href="https://github.com/aptd3v/godock/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=aptd3v/godock" />
</a>

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
		networkoptions.IPAMDriver("default"),
		networkoptions.IPAMConfig("172.20.0.0/16", "", "172.20.0.1"),
		networkoptions.Labels(map[string]string{
			"env": "prod",
			"project": "myapp",
		}),
		networkoptions.EnableIPV6(true),
		networkoptions.Attachable(),
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
	
	volume := volume.NewConfig("my-volume")
	volume.SetOptions(
		// Set the volume driver (with type safety)
		volumeoptions.SetDriver(volumeoptions.NFSDriver),
		
		// Add driver options individually
		volumeoptions.AddDriverOpt("server", "10.0.0.1"),
		volumeoptions.AddDriverOpt("share", "/exports"),
		volumeoptions.AddDriverOpt("security", "sys"),
		
		// Add labels individually
		volumeoptions.AddLabel("environment", "production"),
		volumeoptions.AddLabel("project", "web-app"),
		volumeoptions.AddLabel("team", "backend"),
	)
	
	if err := client.CreateVolume(ctx, volume); err != nil {
		log.Fatal(err)
	}
}

// --- Advanced Volume Configuration ---

// Cluster Volume with Access Control
volume.SetOptions(
	volumeoptions.SetDriver(volumeoptions.LocalDriver),
	volumeoptions.SetClusterSpec(
		"backend-group",
		volumeoptions.SingleNode,
		volumeoptions.ReadWrite,
	),
	volumeoptions.SetCapacityRange(
		100*1024*1024,    // Required: 100MB
		1024*1024*1024,   // Limit: 1GB
	),
	volumeoptions.SetAvailability(volumeoptions.AvailabilityActive),
	
	// Add secrets individually
	volumeoptions.AddSecret("encryption-key", "my-secret-id"),
	volumeoptions.AddSecret("auth-token", "my-auth-secret"),
)
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


### 🔄 godock vs Docker SDK

Here's how godock simplifies Docker operations compared to the raw Docker SDK:

```go
// --- Creating and Starting a Container ---

// Docker SDK
container, err := client.ContainerCreate(ctx,
    &container.Config{
        Image: "nginx",
        ExposedPorts: nat.PortSet{
            "80/tcp": struct{}{},
        },
        Labels: map[string]string{"env": "prod"},
    },
    &container.HostConfig{
        PortBindings: nat.PortMap{
            "80/tcp": []nat.PortBinding{
                {
                    HostIP:   "0.0.0.0",
                    HostPort: "8080",
                },
            },
        },
        RestartPolicy: container.RestartPolicy{
            Name: "always",
        },
        Resources: container.Resources{
            Memory: 256 * 1024 * 1024,
        },
    },
    &network.NetworkingConfig{
        EndpointsConfig: map[string]*network.EndpointSettings{
            "my-net": {
                Aliases: []string{"web-server"},
            },
        },
    },
    "my-nginx",
)

// godock
nginx := container.NewConfig("my-nginx")
nginx.SetContainerOptions(
    containeroptions.Image(nginxImage),
    containeroptions.Expose("80"),
    containeroptions.Label("env", "prod"),
)
nginx.SetHostOptions(
    hostoptions.PortBindings("0.0.0.0", "8080", "80"),
    hostoptions.RestartAlways(),
    hostoptions.Memory(256*1024*1024),
)
nginx.SetNetworkOptions(
    networkoptions.Endpoint("my-net", endpoint),
    networkoptions.Aliases("web-server"),
)

// --- Creating a Network ---

// Docker SDK
_, err := client.NetworkCreate(ctx,
    "my-network",
    types.NetworkCreate{
        Driver: "bridge",
        IPAM: &network.IPAM{
            Driver: "default",
            Config: []network.IPAMConfig{
                {
                    Subnet:  "172.20.0.0/16",
                    Gateway: "172.20.0.1",
                },
            },
        },
        Labels: map[string]string{
            "env": "prod",
            "project": "myapp",
        },
        EnableIPv6: true,
        Internal: false,
        Attachable: true,
    },
)

// godock
net := network.NewConfig("my-network")
net.SetOptions(
    networkoptions.Driver("bridge"),
    networkoptions.IPAMDriver("default"),
    networkoptions.IPAMConfig("172.20.0.0/16", "", "172.20.0.1"),
    networkoptions.Label("env", "prod")
    networkoptions.Label("project", "myapp")
    networkoptions.EnableIPV6(true),
    networkoptions.Attachable(),
)

// --- Network Endpoint Configuration ---

// Docker SDK
container, err := client.ContainerCreate(ctx,
    // ... container config ...
    &network.NetworkingConfig{
        EndpointsConfig: map[string]*network.EndpointSettings{
            "my-net": {
                Aliases: []string{"web-server"},
            },
        },
    },
    // ... other options ...
)

// godock
endpoint := endpointoptions.NewConfig()
endpoint.SetEndpointSetting(
    endpointoptions.Aliases("web-server"),
)
container.SetNetworkOptions(
    networkoptions.Endpoint("my-net", endpoint),
)

// --- Volume Management ---

// Docker SDK
_, err := client.VolumeCreate(ctx,
    volume.CreateOptions{
        Name: "my-volume",
        Driver: "nfs",
        DriverOpts: map[string]string{
            "server": "10.0.0.1",
            "share": "/exports",
            "security": "sys",
        },
        Labels: map[string]string{
            "environment": "production",
            "project": "web-app",
            "team": "backend",
        },
    },
)

// godock
vol := volume.NewConfig("my-volume")
vol.SetOptions(
    // Set the volume driver (with type safety)
    volumeoptions.SetDriver(volumeoptions.NFSDriver),
    
    // Add driver options individually
    volumeoptions.AddDriverOpt("server", "10.0.0.1"),
    volumeoptions.AddDriverOpt("share", "/exports"),
    volumeoptions.AddDriverOpt("security", "sys"),
    
    // Add labels individually
    volumeoptions.AddLabel("environment", "production"),
    volumeoptions.AddLabel("project", "web-app"),
    volumeoptions.AddLabel("team", "backend"),
)

// --- Advanced Volume Configuration ---

// Cluster Volume with Access Control
vol.SetOptions(
    volumeoptions.SetDriver(volumeoptions.LocalDriver),
    volumeoptions.SetClusterSpec(
        "backend-group",
        volumeoptions.SingleNode,
        volumeoptions.ReadWrite,
    ),
    volumeoptions.SetCapacityRange(
        100*1024*1024,    // Required: 100MB
        1024*1024*1024,   // Limit: 1GB
    ),
    volumeoptions.SetAvailability(volumeoptions.AvailabilityActive),
    
    // Add secrets individually
    volumeoptions.AddSecret("encryption-key", "my-secret-id"),
    volumeoptions.AddSecret("auth-token", "my-auth-secret"),
)
```
## Key Differences:
- 📝 **Readability**: Clear, fluent interface vs nested structs and maps
- 🔧 **Maintainability**: Options are grouped logically and self-documenting
- 🚫 **Error Prevention**: Invalid configurations are caught early
- 🎨 **IDE Support**: Better autocompletion and documentation
- 🏗️ **Composability**: Easy to build complex configurations from simple parts



## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
