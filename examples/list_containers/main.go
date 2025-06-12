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
	"github.com/aptd3v/godock/pkg/godock/hostoptions"
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
		err = client.ContainerCreate(ctx, c)
		if err != nil {
			log.Fatalf("failed to create container: %v", err)
		}
		err = client.ContainerStart(ctx, c)
		if err != nil {
			log.Fatalf("failed to start container: %v", err)
		}
	}

	containerList, err := client.ContainerList(ctx,
		godock.WithContainerFilter("status", "running"),
		godock.WithContainerLimit(5),
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
