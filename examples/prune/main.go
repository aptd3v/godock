package main

import (
	"context"
	"fmt"
	"log"

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
	img, err := image.NewConfig("alpine:latest")
	if err != nil {
		log.Fatalf("failed to create image config: %v", err)
	}
	if err := client.PullImage(ctx, img); err != nil {
		log.Fatalf("failed to pull image: %v", err)
	}

	container := container.NewConfig("prune-test")
	container.SetContainerOptions(
		containeroptions.Label("prune-test", "true"),
		containeroptions.CMD("echo", "prune-test"),
	)
	if err := client.CreateContainer(ctx, container); err != nil {
		log.Fatalf("failed to create container: %v", err)
	}

	prune, err := client.ContainerPrune(ctx, godock.WithPruneFilter("label", "prune-test"))
	if err != nil {
		log.Fatalf("failed to prune containers: %v", err)
	}

	fmt.Printf("successfuly pruned containers: %v.\nspace reclaimed: %v\n", prune.ContainersDeleted, prune.SpaceReclaimed)
}
