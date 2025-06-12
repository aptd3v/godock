package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/aptd3v/godock/pkg/godock"
	"github.com/aptd3v/godock/pkg/godock/container"
	"github.com/aptd3v/godock/pkg/godock/containeroptions"
	"github.com/aptd3v/godock/pkg/godock/image"
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

	container := container.NewConfig("prune-test")
	container.SetContainerOptions(
		containeroptions.Label("prune-test", "true"),
		containeroptions.CMD("echo", "prune-test"),
	)
	if err := client.ContainerCreate(ctx, container); err != nil {
		log.Fatalf("failed to create container: %v", err)
	}

	prune, err := client.ContainerPrune(ctx, godock.WithPruneFilter("label", "prune-test"))
	if err != nil {
		log.Fatalf("failed to prune containers: %v", err)
	}

	fmt.Printf("successfully pruned containers: %v.\nspace reclaimed: %v\n", prune.ContainersDeleted, prune.SpaceReclaimed)
}
