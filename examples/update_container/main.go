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
	"github.com/aptd3v/godock/pkg/godock/image"
	"github.com/aptd3v/godock/pkg/godock/updateoptions"
)

func main() {
	ctx := context.Background()
	client, err := godock.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	img, err := image.NewConfig("alpine:latest")
	if err != nil {
		log.Fatalf("failed to create image: %v", err)
	}
	if err := client.PullImage(ctx, img); err != nil {
		log.Fatalf("failed to pull image: %v", err)
	}
	container := container.NewConfig("test-container")
	container.SetContainerOptions(
		containeroptions.Image(img),
		containeroptions.CMD("tail", "-f", "/dev/null"),
	)
	// Setup cleanup on interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanup(client, container)
		os.Exit(1)
	}()
	if err := client.CreateContainer(ctx, container); err != nil {
		log.Fatalf("failed to create container: %v", err)
	}
	if err := client.StartContainer(ctx, container); err != nil {
		log.Fatalf("failed to start container: %v", err)
	}
	sixMegaBytes := int64(6 * 1024 * 1024)
	defer cleanup(client, container)
	warnings, err := client.ContainerUpdate(
		ctx,
		container,
		updateoptions.WithRestartPolicy("no", 0),
		updateoptions.WithMemory(sixMegaBytes),
		updateoptions.WithMemorySwap(sixMegaBytes),
		updateoptions.WithMemoryReservation(sixMegaBytes),
		updateoptions.WithPidsLimit(3),
	)
	if err != nil {
		log.Fatalf("failed to update container: %v", err)
	}
	if len(warnings) > 0 {
		fmt.Println(warnings)
	} else {
		fmt.Println("container updated successfully without warnings")
	}
}
func cleanup(client *godock.Client, container *container.ContainerConfig) {
	ctx := context.Background()
	if err := client.RemoveContainer(ctx, container, true); err != nil {
		log.Fatalf("failed to remove container: %v", err)
	}
}
