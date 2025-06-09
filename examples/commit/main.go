package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aptd3v/godock/pkg/godock"
	"github.com/aptd3v/godock/pkg/godock/commitoptions"
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
		log.Fatalf("failed to create client: %v", err)
	}

	// Setup cleanup on interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	img, err := image.NewConfig("alpine")
	if err != nil {
		log.Fatalf("failed to create image: %v", err)
	}
	if err := client.PullImage(ctx, img); err != nil {
		log.Fatalf("failed to pull image: %v", err)
	}
	commitContainer := container.NewConfig("commit-test")
	commitContainer.SetContainerOptions(
		containeroptions.Image(img),
		containeroptions.CMD("sh", "-c", "apk add --no-cache htop"),
	)
	if err := client.CreateContainer(ctx, commitContainer); err != nil {
		log.Fatalf("failed to create container: %v", err)
	}

	// Start cleanup goroutine
	go func() {
		<-c
		cleanup(client, commitContainer)
		os.Exit(1)
	}()

	if err := client.StartContainer(ctx, commitContainer); err != nil {
		cleanup(client, commitContainer)
		log.Fatalf("failed to start container: %v", err)
	}
	logs, err := client.GetContainerLogs(ctx, commitContainer)
	if err != nil {
		cleanup(client, commitContainer)
		log.Fatalf("failed to get container logs: %v", err)
	}
	//save stdout/stderr to buffer
	stdout := bytes.NewBuffer(nil)
	_, err = godock.NewLogCopier(stdout, os.Stderr).Copy(logs)
	if err != nil {
		cleanup(client, commitContainer)
		log.Fatalf("failed to copy logs: %v", err)
	}
	fmt.Println("stdout", stdout.String())

	commitId, err := client.ContainerCommit(
		ctx,
		commitContainer,
		img,
		commitoptions.Reference("commit-test"),
		commitoptions.Comment("testing commit"),
		commitoptions.Author("aptd3v"),
	)
	if err != nil {
		cleanup(client, commitContainer)
		log.Fatalf("failed to commit container: %v", err)
	}
	fmt.Println("commitId", commitId)

	cleanup(client, commitContainer)

	// Create new container from committed image
	img, err = image.NewConfig(commitId)
	if err != nil {
		log.Fatalf("failed to create image: %v", err)
	}
	commitContainer = container.NewConfig("commit-test-2")
	commitContainer.SetContainerOptions(
		containeroptions.Image(img),
		containeroptions.CMD("tail", "-f", "/dev/null"),
	)
	if err := client.CreateContainer(ctx, commitContainer); err != nil {
		log.Fatalf("failed to create container: %v", err)
	}

	// Update cleanup goroutine for new container
	go func() {
		<-c
		cleanup(client, commitContainer)
		os.Exit(1)
	}()

	if err := client.StartContainer(ctx, commitContainer); err != nil {
		cleanup(client, commitContainer)
		log.Fatalf("failed to start container: %v", err)
	}

	execConfig := exec.NewConfig()
	execConfig.SetOptions(
		execoptions.TTY(true),
		execoptions.AttachStdout(true),
		execoptions.AttachStderr(true),
		execoptions.AttachStdin(true),
		execoptions.CMD("/bin/sh", "-c", "htop"),
	)
	session, err := client.ContainerExecAttachTerminal(ctx, commitContainer, execConfig)
	if err != nil {
		cleanup(client, commitContainer)
		log.Fatalf("failed to attach terminal: %v", err)
	}
	defer session.Close()

	if err := session.Start(); err != nil {
		cleanup(client, commitContainer)
		log.Printf("Session ended with error: %v", err)
	}
	//remove committed container image
	if err := client.RemoveImage(ctx, img, true); err != nil {
		log.Printf("Failed to remove image: %v", err)
	}
}

func cleanup(client *godock.Client, container *container.ContainerConfig) {
	ctx := context.Background()
	fmt.Println("\nCleaning up...")
	if err := client.RemoveContainer(ctx, container, true); err != nil {
		log.Printf("Failed to remove container: %v", err)
	}
}
