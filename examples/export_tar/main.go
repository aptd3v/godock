package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/aptd3v/godock/pkg/godock"
	"github.com/aptd3v/godock/pkg/godock/container"
	"github.com/aptd3v/godock/pkg/godock/containeroptions"
	"github.com/aptd3v/godock/pkg/godock/hostoptions"
	"github.com/aptd3v/godock/pkg/godock/image"
)

func main() {
	ctx := context.Background()
	client, err := godock.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	image, err := image.NewConfig("alpine:latest")
	if err != nil {
		log.Fatalf("failed to create image: %v", err)
	}
	if err := client.PullImage(ctx, image); err != nil {
		log.Fatalf("failed to pull image: %v", err)
	}
	container := container.NewConfig("export-container-tar-test")
	container.SetContainerOptions(
		containeroptions.Image(image),
		//long running command to test export
		containeroptions.CMD("tail", "-f", "/dev/null"),
	)
	container.SetHostOptions(
		hostoptions.AutoRemove(),
	)
	if err := client.CreateContainer(ctx, container); err != nil {
		log.Fatalf("failed to create container: %v", err)
	}
	if err := client.StartContainer(ctx, container); err != nil {
		log.Fatalf("failed to start container: %v", err)
	}
	export, err := client.ContainerExport(ctx, container)
	if err != nil {
		log.Fatalf("failed to export container: %v", err)
	}
	defer export.Close()

	//save to file
	file, err := os.Create("export-example.tar")
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, export); err != nil {
		log.Fatalf("failed to write export: %v", err)
	}

	//load the image from the exported tar file
	exec.Command("docker", "load", "-i", "export.tar").Run()

	if err := client.RemoveContainer(ctx, container, true); err != nil {
		log.Fatalf("failed to remove container: %v", err)
	}
}
