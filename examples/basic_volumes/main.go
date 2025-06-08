package main

import (
	"context"
	"log"

	"github.com/aptd3v/godock/pkg/godock"
	"github.com/aptd3v/godock/pkg/godock/volume"
	"github.com/aptd3v/godock/pkg/godock/volumeoptions"
)

func main() {
	ctx := context.Background()
	client, err := godock.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	volume := volume.NewConfig("my-data")
	volume.SetOptions(
		volumeoptions.Driver("local"),
		volumeoptions.Label("backup", "true"),
	)

	if err := client.CreateVolume(ctx, volume); err != nil {
		log.Fatal(err)
	}
}
