package main

import (
	"context"
	"fmt"
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
	ubuntuImg, err := image.NewConfig("ubuntu:22.04")
	if err != nil {
		log.Fatal(err)
	}
	err = client.PullImage(ctx, ubuntuImg)
	if err != nil {
		log.Fatal(err)
	}

	// Create container config
	containerConfig := container.NewConfig("exec-example")
	containerConfig.SetContainerOptions(
		containeroptions.Image(ubuntuImg),
		containeroptions.CMD("tail", "-f", "/dev/null"), // Keep container running
	)

	// Set up cleanup on interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\nCleaning up...")
		cleanup(client, containerConfig)
		os.Exit(0)
	}()
	defer cleanup(client, containerConfig)

	// Create and start the container
	fmt.Println("Creating and starting container...")
	if err := client.CreateContainer(ctx, containerConfig); err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}

	if err := client.StartContainer(ctx, containerConfig); err != nil {
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
	session, err := client.ContainerExecAttachTerminal(ctx, containerConfig, execConfig)
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

	if err := client.StopContainer(ctx, containerConfig); err != nil {
		log.Printf("Failed to stop container: %v", err)
	}

	if err := client.RemoveContainer(ctx, containerConfig, true); err != nil {
		log.Printf("Failed to remove container: %v", err)
	}

	fmt.Println("Cleanup completed")
}
