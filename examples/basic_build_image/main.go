package main

import (
	"context"
	"log"

	"github.com/aptd3v/godock/pkg/godock"
	"github.com/aptd3v/godock/pkg/godock/image"
)

func main() {
	ctx := context.Background()
	client, err := godock.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Create image config from Dockerfile in current directory
	img, err := image.NewImageFromSrc("./examples/basic_build_image")
	if err != nil {
		log.Fatal(err)
	}

	// Build the image
	if err := client.BuildImage(ctx, img); err != nil {
		log.Fatal(err)
	}
}
