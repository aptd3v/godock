package image

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aptd3v/godock/pkg/godock/imageoptions"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
)

// Image represents a Docker image and provides methods for setting pull and build options.
type ImageConfig struct {
	Ref          string
	BuildOptions *types.ImageBuildOptions
	PullOptions  *image.PullOptions
	PushOptions  *image.PushOptions
}

// SetPullOptions configures pull options for the Docker image.
// Use this method to set various pull options using functions from the imageoptions package.
func (img *ImageConfig) SetPullOptions(setOFns ...imageoptions.SetPullOptFn) {
	for _, set := range setOFns {
		set(img.PullOptions)
	}
}

// SetPushOptions configures push options for the Docker image.
// Use this method to set various push options using functions from the imageoptions package.
func (img *ImageConfig) SetPushOptions(setOFns ...imageoptions.SetPushOptFn) {
	for _, set := range setOFns {
		set(img.PushOptions)
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

// ValidateReference checks if the image reference is valid.
// It ensures the reference follows Docker's image naming conventions.
func ValidateReference(ref string) error {
	if ref == "" {
		return fmt.Errorf("image reference cannot be empty")
	}

	// Basic format validation
	parts := strings.Split(ref, "/")
	if len(parts) > 3 {
		return fmt.Errorf("invalid image reference format: %s", ref)
	}

	// Check for valid registry format if present
	if len(parts) > 2 {
		registry := parts[0]
		if !strings.Contains(registry, ".") && registry != "localhost" {
			return fmt.Errorf("invalid registry format in reference: %s", ref)
		}
	}

	// Check tag/digest format
	name := parts[len(parts)-1]
	if strings.Count(name, ":") > 1 {
		return fmt.Errorf("invalid tag format in reference: %s", ref)
	}

	if strings.Contains(name, "@") && !strings.Contains(name, "sha256:") {
		return fmt.Errorf("invalid digest format in reference: %s", ref)
	}

	return nil
}

// NewConfig creates a new Image configuration with the specified image reference.
// The Image instance contains pull, push, and build options for the Docker image.
func NewConfig(ref string) (*ImageConfig, error) {
	if err := ValidateReference(ref); err != nil {
		return nil, err
	}

	return &ImageConfig{
		Ref:          ref,
		BuildOptions: &types.ImageBuildOptions{},
		PullOptions:  &image.PullOptions{},
		PushOptions:  &image.PushOptions{},
	}, nil
}

/*
NewImageFromSrc creates a new Image configuration from a source directory.
The directory must contain a Dockerfile in its root.
This is equivalent to running `docker build` with the specified directory as context.

Usage example:

	img, err := image.NewImageFromSrc("./myapp")
	if err != nil {
		return err
	}
	img.SetBuildOptions(
		imageoptions.SetTag("myapp:latest"),
		imageoptions.AddBuildArg("VERSION", "1.0.0"),
	)
*/
func NewImageFromSrc(dir string) (*ImageConfig, error) {
	context, err := createLocalBuildContext(dir)
	if err != nil {
		return nil, err
	}

	// Check for Dockerfile
	if _, err := os.Stat(filepath.Join(dir, "Dockerfile")); os.IsNotExist(err) {
		return nil, fmt.Errorf("Dockerfile not found in directory: %s", dir)
	}

	return &ImageConfig{
		Ref: "",
		BuildOptions: &types.ImageBuildOptions{
			Context: context,
		},
		PullOptions: &image.PullOptions{},
		PushOptions: &image.PushOptions{},
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
