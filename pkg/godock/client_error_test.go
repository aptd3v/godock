package godock

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/aptd3v/godock/pkg/godock/container"
	"github.com/aptd3v/godock/pkg/godock/containeroptions"
	"github.com/aptd3v/godock/pkg/godock/errdefs"
	"github.com/aptd3v/godock/pkg/godock/hostoptions"
	"github.com/aptd3v/godock/pkg/godock/image"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClienterrdefs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	ctx := context.Background()
	client, err := NewClient(ctx)
	require.NoError(t, err)

	// Cleanup any leftover test containers
	cleanup := func(prefix string) {
		containers, err := client.ContainerList(ctx, WithContainerAll(true))
		if err != nil {
			return
		}
		for _, c := range containers {
			for _, name := range c.Names {
				if len(name) > 1 && name[1:] == prefix {
					// Stop container first if running
					stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
					client.ContainerStop(stopCtx, &container.ContainerConfig{Id: c.ID})
					cancel()

					// Remove container with force
					removeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
					client.ContainerRemove(removeCtx, &container.ContainerConfig{Id: c.ID}, true)
					cancel()

					// Wait a bit to ensure removal is complete
					time.Sleep(1 * time.Second)
				}
			}
		}
	}

	// Pull required images
	pullImage := func(ref string) error {
		imgConfig := image.NewConfig(ref)
		rc, err := client.ImagePull(ctx, imgConfig)
		if err != nil {
			return err
		}
		defer rc.Close()

		// Copy output to discard to ensure pull completes
		_, err = io.Copy(os.Stdout, rc)
		if err != nil {
			return err
		}

		// Wait a bit to ensure pull is complete
		time.Sleep(2 * time.Second)
		return nil
	}

	// Pull required images upfront
	require.NoError(t, pullImage("alpine:latest"))
	require.NoError(t, pullImage("nginx:latest"))

	t.Run("NotFound/Image", func(t *testing.T) {
		config := container.NewConfig("test-not-found")
		config.Options.Image = "nonexistent:latest"

		err := client.ContainerCreate(ctx, config)
		require.Error(t, err)
		assert.ErrorIs(t, err, errdefs.ErrNotFound)

		var nfe *errdefs.ResourceNotFoundError
		require.ErrorAs(t, err, &nfe)
		assert.Equal(t, "image", nfe.ResourceType)
		assert.Equal(t, "nonexistent:latest", nfe.ID)
	})

	t.Run("AlreadyExists/ContainerName", func(t *testing.T) {
		cleanup("test-exists")

		// Create first container
		config := container.NewConfig("test-exists")
		config.Options.Image = "alpine:latest"
		err := client.ContainerCreate(ctx, config)
		require.NoError(t, err)
		defer func() {
			client.ContainerStop(ctx, config)
			client.ContainerRemove(ctx, config, true)
		}()

		// Try to create container with same name
		config2 := container.NewConfig("test-exists")
		config2.Options.Image = "alpine:latest"
		err = client.ContainerCreate(ctx, config2)
		require.Error(t, err)
		assert.ErrorIs(t, err, errdefs.ErrAlreadyExists)

		var ee *errdefs.ResourceExistsError
		require.ErrorAs(t, err, &ee)
		assert.Equal(t, "container", ee.ResourceType)
		assert.Equal(t, "test-exists", ee.ID)
	})

	t.Run("AlreadyExists/PortConflict", func(t *testing.T) {
		cleanup("test-port1")
		cleanup("test-port2")
		time.Sleep(2 * time.Second) // Extra wait to ensure cleanup

		// Create and start first container with port 8080
		config := container.NewConfig("test-port1")
		config.Options.Image = "nginx:latest"
		config.SetContainerOptions(
			containeroptions.Image(image.NewConfig("nginx:latest")),
			containeroptions.Expose("80/tcp"),
		)
		config.SetHostOptions(hostoptions.PortBindings("0.0.0.0", "8080", "80/tcp"))
		err := client.ContainerCreate(ctx, config)
		require.NoError(t, err)
		err = client.ContainerStart(ctx, config)
		require.NoError(t, err)
		defer func() {
			client.ContainerStop(ctx, config)
			client.ContainerRemove(ctx, config, true)
		}()

		time.Sleep(2 * time.Second) // Wait for container to fully start

		// Try to create and start second container with same port
		config2 := container.NewConfig("test-port2")
		config2.Options.Image = "nginx:latest"
		config2.SetContainerOptions(
			containeroptions.Image(image.NewConfig("nginx:latest")),
			containeroptions.Expose("80/tcp"),
		)
		config2.SetHostOptions(hostoptions.PortBindings("0.0.0.0", "8080", "80/tcp"))
		err = client.ContainerCreate(ctx, config2)
		require.NoError(t, err) // Creation should succeed
		defer client.ContainerRemove(ctx, config2, true)

		err = client.ContainerStart(ctx, config2)
		require.Error(t, err)
		assert.ErrorIs(t, err, errdefs.ErrAlreadyExists)

		var ee *errdefs.ResourceExistsError
		require.ErrorAs(t, err, &ee)
		assert.Equal(t, "port", ee.ResourceType)
		assert.Equal(t, "test-port2", ee.ID)
	})

	t.Run("InvalidConfig/NilConfig", func(t *testing.T) {
		err := client.ContainerCreate(ctx, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, errdefs.ErrInvalidConfig)

		var ve *errdefs.ValidationError
		require.ErrorAs(t, err, &ve)
		assert.Equal(t, "containerConfig", ve.Field)
		assert.Equal(t, "container config cannot be nil", ve.Message)
	})

	t.Run("ContainerError/InvalidCommand", func(t *testing.T) {
		cleanup("test-cmd")
		time.Sleep(2 * time.Second) // Extra wait to ensure cleanup

		config := container.NewConfig("test-cmd")
		config.Options.Image = "alpine:latest"
		config.Options.Cmd = []string{"/nonexistent"}
		err := client.ContainerCreate(ctx, config)
		require.NoError(t, err)
		defer func() {
			client.ContainerStop(ctx, config)
			client.ContainerRemove(ctx, config, true)
		}()

		err = client.ContainerStart(ctx, config)
		require.Error(t, err)

		var ce *errdefs.ContainerError
		require.ErrorAs(t, err, &ce)
		assert.Equal(t, "test-cmd", ce.ID)
		assert.Equal(t, "start", ce.Op)
	})

	t.Run("Timeout/ContainerWait", func(t *testing.T) {
		cleanup("test-timeout")
		time.Sleep(2 * time.Second) // Extra wait to ensure cleanup

		config := container.NewConfig("test-timeout")
		config.Options.Image = "alpine:latest"
		config.Options.Cmd = []string{"sleep", "30"}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := client.RunAndWait(ctx, config)
		require.Error(t, err)
		assert.ErrorIs(t, err, errdefs.ErrTimeout)

		cleanup("test-timeout") // Clean up after test
	})

	t.Run("Canceled/ContainerWait", func(t *testing.T) {
		cleanup("test-cancel")
		time.Sleep(2 * time.Second) // Extra wait to ensure cleanup

		config := container.NewConfig("test-cancel")
		config.Options.Image = "alpine:latest"
		config.Options.Cmd = []string{"sleep", "30"}

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(1 * time.Second)
			cancel()
		}()

		err := client.RunAndWait(ctx, config)
		require.Error(t, err)
		assert.ErrorIs(t, err, errdefs.ErrCanceled)

		cleanup("test-cancel") // Clean up after test
	})
}
