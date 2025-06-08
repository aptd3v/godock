package main

import (
	"context"
	"log"

	"github.com/aptd3v/godock/pkg/godock"
	"github.com/aptd3v/godock/pkg/godock/network"
	"github.com/aptd3v/godock/pkg/godock/networkoptions"
)

func main() {
	ctx := context.Background()
	client, err := godock.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	network := network.NewConfig("my-network")
	network.SetOptions(
		networkoptions.Driver("bridge"),
		networkoptions.Label("environment", "development"),
		networkoptions.EnableIPV6(true),
	)

	if err := client.CreateNetwork(ctx, network); err != nil {
		log.Fatal(err)
	}
}
