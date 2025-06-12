package main

import (
	"context"
	"io"
	"log"
	"os"

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
	rc, err := client.ImageBuild(ctx, img)
	if err != nil {
		log.Fatal(err, "failed to build image")
	}
	defer rc.Close()
	_, err = io.Copy(os.Stdout, rc)
	if err != nil {
		log.Fatal(err, "failed to copy logs")
	}
}
