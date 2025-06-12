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
	"github.com/aptd3v/godock/pkg/godock/image"
	"github.com/aptd3v/godock/pkg/godock/updateoptions"
)

func main() {
	ctx := context.Background()
	client, err := godock.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	img := image.NewConfig("alpine")
	rc, err := client.ImagePull(ctx, img)
	if err != nil {
		log.Fatalf("failed to pull image: %v", err)
	}
	defer rc.Close()
	_, err = io.Copy(os.Stdout, rc)
	if err != nil {
		log.Fatalf("failed to copy logs: %v", err)
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
	if err := client.ContainerCreate(ctx, container); err != nil {
		log.Fatalf("failed to create container: %v", err)
	}
	if err := client.ContainerStart(ctx, container); err != nil {
		log.Fatalf("failed to start container: %v", err)
	}
	sixMegaBytes := int64(6 * 1024 * 1024)
	defer cleanup(client, container)
	res, err := client.ContainerUpdate(
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
	if len(res.Warnings) > 0 {
		fmt.Println(res.Warnings)
	} else {
		fmt.Println("container updated successfully without warnings")
	}
}
func cleanup(client *godock.Client, container *container.ContainerConfig) {
	ctx := context.Background()
	if err := client.ContainerRemove(ctx, container, true); err != nil {
		log.Fatalf("failed to remove container: %v", err)
	}
}
