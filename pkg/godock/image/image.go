package image

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aptd3v/godock/pkg/godock/imageoptions"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
)

// Image represents a Docker image and provides methods for setting pull and build options.
type ImageConfig struct {
	Ref          string
	BuildOptions *types.ImageBuildOptions
	PullOptions  *image.PullOptions
}

// SetPullOptions configures pull options for the Docker image.
// Use this method to set various pull options using functions from the imageoptions package.
func (img *ImageConfig) SetPullOptions(setOFns ...imageoptions.SetPullOptFn) {
	for _, set := range setOFns {
		set(img.PullOptions)
	}
}

// SetBuildOptions configures build options for the Docker image.
// Use this method to set various build options using functions from the imageoptions package.
func (img *ImageConfig) SetBuildOptions(setOFns ...imageoptions.SetBuildOptFn) {
	for _, set := range setOFns {
		set(img.BuildOptions)
	}
}

// String returns the reference of the Docker image.
func (img *ImageConfig) String() string {
	return img.Ref
}

// NewConfig creates a new Image configuration with the specified image reference.
// The Image instance contains pull and build options for the Docker image.
func NewConfig(ref string) *ImageConfig {
	return &ImageConfig{
		Ref:          ref,
		BuildOptions: &types.ImageBuildOptions{},
		PullOptions:  &image.PullOptions{},
	}
}

/*
Build from a Dockerfile in the provided directory,
ensure that the Dockerfile exists in the root path
of that directory for this function to work correctly.
*/
func NewImageFromSrc(dir string) (*ImageConfig, error) {
	context, err := createLocalBuildContext(dir)
	if err != nil {
		return nil, err
	}
	return &ImageConfig{
		Ref: "",
		BuildOptions: &types.ImageBuildOptions{
			Context: context,
		},
	}, nil
}

// Archives a directory for docker build context
func createLocalBuildContext(src string) (io.ReadCloser, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Ensure sourceDir exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil, fmt.Errorf("source directory %s does not exist", src)
	}

	// Walk through the source directory and add files to the tar archive
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the source directory itself
		if path == src {
			return nil
		}

		// Create a tar header from the file info
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		// Set the correct path for the file in the archive
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		// Write the tar header and file content to the tar writer
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tw, file)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	return io.NopCloser(&buf), nil
}
