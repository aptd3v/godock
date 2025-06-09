# GoDock Examples

This directory contains examples demonstrating various features of the GoDock library.

## Examples

### Basic Examples

#### Basic Container (`basic_container/main.go`)
Shows basic container operations:
- Creating and running containers
- Basic container lifecycle management
- Error handling

#### Basic Network (`basic_network/main.go`)
Demonstrates network operations:
- Creating custom networks
- Connecting containers to networks
- Network configuration

#### Basic Volumes (`basic_volumes/main.go`)
Shows volume management:
- Creating and managing volumes
- Mounting volumes to containers
- Volume data persistence

#### Basic Build Image (`basic_build_image/main.go`)
Demonstrates image building:
- Building custom images
- Using Dockerfiles
- Image management

### Advanced Examples

#### Stats Example (`stats/main.go`)
Shows how to monitor container resource usage in real-time:
- CPU usage
- Memory usage
- Network I/O
- Interactive shell access while monitoring

#### Commit Example (`commit/main.go`)
Demonstrates container image manipulation:
- Creating a container from base image
- Installing software (htop)
- Committing changes to a new image
- Running an interactive terminal with the new image
- Proper cleanup of containers and images

#### List Containers (`list_containers/main.go`)
Shows container listing and filtering:
- Listing running containers
- Filtering containers by status
- Container information display

#### Export TAR (`export_tar/main.go`)
Demonstrates container export functionality:
- Exporting containers to tar archives
- Container filesystem manipulation

#### Exec Example (`exec/main.go`)
Shows container command execution:
- Running commands in containers
- Interactive terminal sessions
- Command output handling

#### Container API (`container_api/main.go`)
Demonstrates the container API:
- Container configuration
- Container lifecycle management
- API usage patterns

### Update Container Example
Demonstrates the container API:
- Container update management

### Application Examples

#### Web App (`webapp/main.go`)
Shows how to run a web application:
- Web server container setup
- Port mapping
- Web application deployment

#### MongoDB (`mongodb/main.go`)
Demonstrates running MongoDB:
- Database container setup
- Data persistence
- Container networking

#### Redis (`redis/main.go`)
Shows Redis container management:
- Redis server setup
- Redis configuration
- Data persistence

## Running Examples

1. Make sure Docker is running on your system
2. Clone the repository
3. Navigate to the project root
4. Run any example using `go run examples/<example>/main.go`

## Common Patterns

All examples follow these best practices:
- Use of context for cancellation
- Proper resource cleanup on exit
- Signal handling for graceful shutdown
- Error handling at each step
- Clear logging of operations

## Contributing

When adding new examples:
1. Create a new directory under `examples/`
2. Follow the existing pattern for error handling and cleanup
3. Add documentation to this README
4. Include comments in your code explaining key concepts
5. Ensure proper resource cleanup
6. Add appropriate error handling
7. Include signal handling for graceful shutdown 