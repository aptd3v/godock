package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/aptd3v/godock/pkg/godock"
	"github.com/aptd3v/godock/pkg/godock/container"
	"github.com/aptd3v/godock/pkg/godock/containeroptions"
	"github.com/aptd3v/godock/pkg/godock/hostoptions"
	"github.com/aptd3v/godock/pkg/godock/image"
	"github.com/google/uuid"
)

// ContainerRequest represents the JSON structure for container creation
type ContainerRequest struct {
	Name    string   `json:"name,omitempty"`
	Image   string   `json:"image"`
	Command []string `json:"command,omitempty"`
}

// ContainerResponse represents the API response
type ContainerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type API struct {
	client *godock.Client
}

func (a *API) runContainer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ContainerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Setup container
	name := req.Name + "_" + uuid.NewString()
	img := image.NewConfig(req.Image)

	// Pull image silently
	if _, err := a.client.ImagePull(context.Background(), img); err != nil {
		http.Error(w, fmt.Sprintf("Failed to pull image: %v", err), http.StatusInternalServerError)
		return
	}

	// Create and configure container
	c := container.NewConfig(name)
	opts := []containeroptions.SetOptionsFns{
		containeroptions.Image(img),
		containeroptions.AttachStdout(),
		containeroptions.AttachStderr(),
	}
	if len(req.Command) > 0 {
		opts = append(opts, containeroptions.CMD(req.Command...))
	}
	c.SetContainerOptions(opts...)
	c.SetHostOptions(hostoptions.AutoRemove())

	// Create and start container
	if err := a.client.ContainerCreate(r.Context(), c); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create container: %v", err), http.StatusInternalServerError)
		return
	}
	if err := a.client.ContainerStart(r.Context(), c); err != nil {
		http.Error(w, fmt.Sprintf("Failed to start container: %v", err), http.StatusInternalServerError)
		return
	}

	// Stream logs
	w.Header().Set("Content-Type", "text/plain")
	logs, err := a.client.ContainerLogs(r.Context(), c)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get logs: %v", err), http.StatusInternalServerError)
		return
	}

	// Use LogCopier with prefixes to distinguish stdout and stderr
	copier := godock.NewLogCopier(os.Stdout, nil)
	if _, err := copier.CopyWithPrefix(logs, "[stdout] ", "[stderr] "); err != nil {
		log.Printf("Error copying logs: %v", err)
	}
}

func main() {
	client, err := godock.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}

	api := &API{client: client}
	http.HandleFunc("/containers", api.runContainer)

	srv := &http.Server{Addr: ":5000"}
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		srv.Shutdown(context.Background())
	}()

	log.Printf("Server starting on :5000")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

/*
To test the container API, run the following commands:

go run examples/container_api/main.go

# Run a simple command in a container
curl -X POST http://localhost:5000/containers \
  -H "Content-Type: text/plain" \
  -d '{
    "name": "test",
    "image": "alpine:latest",
    "command": ["sh", "-c", "echo hello && echo error >&2"]
  }'
*/
