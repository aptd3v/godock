package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aptd3v/godock/pkg/godock"
	"github.com/aptd3v/godock/pkg/godock/container"
	"github.com/aptd3v/godock/pkg/godock/containeroptions"
	"github.com/aptd3v/godock/pkg/godock/hostoptions"
	"github.com/aptd3v/godock/pkg/godock/image"
	"github.com/aptd3v/godock/pkg/godock/listoptions"
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
	err = client.PullImage(ctx, img)
	if err != nil {
		log.Fatalf("failed to pull image: %v", err)
	}
	containerConfigs := make([]*container.ContainerConfig, 10)
	for i := 0; i < 10; i++ {
		c := container.NewConfig(fmt.Sprintf("container-list-test-%d", i))
		c.SetContainerOptions(
			containeroptions.Image(img),
			containeroptions.CMD("sleep", fmt.Sprintf("%d", i)),
		)
		c.SetHostOptions(
			hostoptions.AutoRemove(),
		)
		containerConfigs[i] = c
	}
	for _, c := range containerConfigs {
		err = client.CreateContainer(ctx, c)
		if err != nil {
			log.Fatalf("failed to create container: %v", err)
		}
		err = client.StartContainer(ctx, c)
		if err != nil {
			log.Fatalf("failed to start container: %v", err)
		}
	}

	containerList, err := client.ContainerList(ctx,
		listoptions.WithFilters("status", "running"),
		listoptions.WithLimit(5),
	)
	if err != nil {
		log.Fatalf("failed to list containers: %v", err)
	}
	for _, c := range containerList {
		fmt.Printf("Container: %s\n", c.Names[0])
		fmt.Printf("  ID: %s\n", c.ID[:12])
		fmt.Printf("  Image: %s\n", c.Image)
		fmt.Printf("  Command: %s\n", c.Command)
		fmt.Printf("  Status: %s\n", c.Status)
		fmt.Printf("  Size: %d bytes\n", c.SizeRw)
	}
}
