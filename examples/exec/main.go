package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aptd3v/godock/pkg/godock"
	"github.com/aptd3v/godock/pkg/godock/container"
	"github.com/aptd3v/godock/pkg/godock/containeroptions"
	"github.com/aptd3v/godock/pkg/godock/exec"
	"github.com/aptd3v/godock/pkg/godock/execoptions"
	"github.com/aptd3v/godock/pkg/godock/image"
)

func main() {
	ctx := context.Background()
	client, err := godock.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Pull Ubuntu image
	ubuntuImage := image.NewConfig("ubuntu")
	fmt.Println(ubuntuImage.Ref)
	rc, err := client.ImagePull(ctx, ubuntuImage)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(os.Stdout, rc)
	if err != nil {
		log.Fatal(err)
	}

	// Create container config
	ubuntuContainer := container.NewConfig("exec-example")
	ubuntuContainer.SetContainerOptions(
		containeroptions.Image(ubuntuImage),
		containeroptions.CMD("tail", "-f", "/dev/null"), // Keep container running
	)

	// Set up cleanup on interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nCleaning up...")
		cleanup(client, ubuntuContainer)
		os.Exit(0)
	}()
	defer cleanup(client, ubuntuContainer)

	// Create and start the container
	fmt.Println("Creating and starting container...")
	if err := client.ContainerCreate(ctx, ubuntuContainer); err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}

	if err := client.ContainerStart(ctx, ubuntuContainer); err != nil {
		log.Fatalf("Failed to start container: %v", err)
	}

	// Create exec config for interactive shell
	execConfig := exec.NewConfig()
	execConfig.SetOptions(
		execoptions.TTY(true),
		execoptions.AttachStdin(true),
		execoptions.AttachStdout(true),
		execoptions.AttachStderr(true),
		execoptions.CMD("/bin/bash"),
	)

	// Execute command in container
	fmt.Println("\nStarting interactive shell...")
	session, err := client.ContainerExecAttachTerminal(ctx, ubuntuContainer, execConfig)
	if err != nil {
		log.Fatalf("Failed to execute command: %v", err)
	}
	defer session.Close()

	// Start the interactive session
	if err := session.Start(); err != nil {
		log.Printf("Session ended with error: %v", err)
	}
}

func cleanup(client *godock.Client, containerConfig *container.ContainerConfig) {
	ctx := context.Background()

	if err := client.ContainerStop(ctx, containerConfig); err != nil {
		log.Printf("Failed to stop container: %v", err)
	}

	if err := client.ContainerRemove(ctx, containerConfig, true); err != nil {
		log.Printf("Failed to remove container: %v", err)
	}

	fmt.Println("Cleanup completed")
}
