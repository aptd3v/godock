package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aptd3v/godock/pkg/godock"
	"github.com/aptd3v/godock/pkg/godock/container"
	"github.com/aptd3v/godock/pkg/godock/containeroptions"
	"github.com/aptd3v/godock/pkg/godock/image"
)

// ContainerRequest represents the JSON structure for container creation
type ContainerRequest struct {
	Name          string            `json:"name"`
	Image         string            `json:"image"`
	Ports         []PortMapping     `json:"ports"`
	Env           map[string]string `json:"env"`
	Labels        map[string]string `json:"labels"`
	Memory        int64             `json:"memory_bytes,omitempty"`
	CPUQuota      int64             `json:"cpu_quota_microseconds,omitempty"` // CPU quota in microseconds (minimum 1000)
	RestartPolicy string            `json:"restart_policy,omitempty"`
	Command       []string          `json:"command,omitempty"`
}

// PortMapping represents a port mapping configuration
type PortMapping struct {
	HostIP        string `json:"host_ip"`
	HostPort      string `json:"host_port"`
	ContainerPort string `json:"container_port"`
}

// ContainerResponse represents the API response
type ContainerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type containerHandler struct {
	client *godock.Client
}

func (h *containerHandler) createContainer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if we have markdown content
	var markdownContent []byte
	if r.Header.Get("Content-Type") == "text/markdown" {
		var err error
		markdownContent, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read markdown content", http.StatusBadRequest)
			return
		}
		// Create a new reader for the JSON config that follows
		r.Body = io.NopCloser(bytes.NewReader(markdownContent))
	}

	// Set headers for binary response
	w.Header().Set("Content-Type", "application/pdf")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	var req ContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Fprintf(w, "Error: Invalid request body: %v\n", err)
		return
	}

	// Create cleanup function to handle errors
	cleanup := func(c *container.ContainerConfig, err error) {
		if c != nil {
			_ = h.client.RemoveContainer(r.Context(), c, true)
		}
		if err != nil {
			fmt.Fprintf(w, "Error: %v\n", err)
		}
	}

	// Create a temporary container config just for cleanup
	cleanupContainer := container.NewConfig(req.Name)
	// Try to remove any existing container with the same name
	// Ignore errors since the container might not exist
	_ = h.client.RemoveContainer(r.Context(), cleanupContainer, true)

	// Create image config
	img, err := image.NewConfig(req.Image)
	if err != nil {
		cleanup(nil, fmt.Errorf("invalid image configuration: %v", err))
		return
	}

	// Pull the image
	if err := h.client.PullImage(r.Context(), img); err != nil {
		cleanup(nil, fmt.Errorf("failed to pull image: %v", err))
		return
	}

	// Create container config
	c := container.NewConfig(req.Name)

	// Container options
	containerOpts := []containeroptions.SetOptionsFns{
		containeroptions.Image(img),
		containeroptions.AttachStdout(),
		containeroptions.AttachStderr(),
	}

	// Add command if specified
	if len(req.Command) > 0 {
		containerOpts = append(containerOpts, containeroptions.CMD(req.Command...))
	}

	c.SetContainerOptions(containerOpts...)

	// Create and start the container
	if err := h.client.CreateContainer(r.Context(), c); err != nil {
		cleanup(c, fmt.Errorf("failed to create container: %v", err))
		return
	}

	// Set up writer that writes to response and flushes
	writer := &flushWriter{w: w, f: flusher}
	h.client.SetLogResponseWriter(writer)
	h.client.SetImageResponeWriter(writer)

	// Start container
	if err := h.client.StartContainer(r.Context(), c); err != nil {
		cleanup(c, fmt.Errorf("failed to start container: %v", err))
		return
	}

	// Start streaming logs immediately in a goroutine
	logErrCh := make(chan error, 1)
	go func() {
		logErrCh <- h.client.GetContainerLogs(r.Context(), c)
	}()

	// Wait for container to finish
	statusCh, errCh := h.client.ContainerWait(r.Context(), c)
	select {
	case err := <-errCh:
		cleanup(c, fmt.Errorf("container failed: %v", err))
		return
	case <-statusCh:
		// Wait for logs to finish streaming
		if err := <-logErrCh; err != nil {
			cleanup(c, fmt.Errorf("failed to get logs: %v", err))
			return
		}
	}

	// Clean up container
	if err := h.client.RemoveContainer(r.Context(), c, true); err != nil {
		fmt.Fprintf(w, "Warning: Failed to remove container: %v\n", err)
	}
}

// flushWriter wraps a ResponseWriter to provide automatic flushing
type flushWriter struct {
	w http.ResponseWriter
	f http.Flusher
}

func (fw *flushWriter) Write(p []byte) (n int, err error) {
	if fw.w == nil {
		return 0, fmt.Errorf("response writer is nil")
	}
	n, err = fw.w.Write(p)
	if err == nil && fw.f != nil {
		fw.f.Flush()
	}
	return
}

func main() {
	// Create a context that we can cancel on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Docker client
	client, err := godock.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	// Create handler
	handler := &containerHandler{client: client}

	// Set up HTTP server
	http.HandleFunc("/containers", handler.createContainer)

	// Handle shutdown gracefully
	server := &http.Server{
		Addr: ":5000",
	}

	// Handle interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		server.Shutdown(ctx)
	}()

	log.Printf("Server starting on :5000")
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

/*
To test the container API, run the following commands:

go run examples/container_api/main.go

# Run a simple command in a container
curl -X POST http://localhost:5000/containers \
  --no-buffer \
  --output - \
  -H "Content-Type: application/json" \
  -d '{
    "name": "hello",
    "image": "alpine:latest",
    "command": ["echo", "Hello from container!"]
  }'
*/
