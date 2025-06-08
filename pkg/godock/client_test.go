package godock

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/aptd3v/godock/pkg/godock/container"
	"github.com/aptd3v/godock/pkg/godock/image"
	"github.com/aptd3v/godock/pkg/godock/network"
	"github.com/aptd3v/godock/pkg/godock/volume"
	"github.com/stretchr/testify/require"
)

func setupTestClient(t *testing.T) *Client {
	ctx := context.Background()
	client, err := NewClient(ctx)
	if err != nil {
		t.Skipf("Docker daemon is not running: %v", err)
	}
	return client
}

func setupTestImage(ctx context.Context, t *testing.T, client *Client) *image.ImageConfig {
	imageConfig, err := image.NewConfig("alpine:latest")
	require.NoError(t, err)

	err = client.PullImage(ctx, imageConfig)
	require.NoError(t, err)

	return imageConfig
}

func TestClientCreation(t *testing.T) {
	ctx := context.Background()
	client, err := NewClient(ctx)
	if err != nil {
		t.Skipf("Docker daemon is not running: %v", err)
	}
	require.NotNil(t, client)
	require.NotNil(t, client.wrapped)
	require.NotNil(t, client.imageResWriter)
	require.NotNil(t, client.statsResWriter)
	require.NotNil(t, client.logResWriter)
}

func TestContainerLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)
	imageConfig := setupTestImage(ctx, t, client)

	t.Run("Basic Container Operations", func(t *testing.T) {
		containerConfig := container.NewConfig("test-container")
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"sleep", "1"}

		// Create container
		err := client.CreateContainer(ctx, containerConfig)
		require.NoError(t, err)
		require.NotEmpty(t, containerConfig.Id)
		defer client.RemoveContainer(ctx, containerConfig, true)

		// Start container
		err = client.StartContainer(ctx, containerConfig)
		require.NoError(t, err)

		// Wait for container with timeout
		waitCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		waitCh, errCh := client.ContainerWait(waitCtx, containerConfig)
		select {
		case result := <-waitCh:
			require.Equal(t, int64(0), result.StatusCode)
		case err := <-errCh:
			require.NoError(t, err)
		case <-waitCtx.Done():
			t.Fatal("Container wait timed out")
		}
	})

	t.Run("Container Pause and Unpause", func(t *testing.T) {
		containerConfig := container.NewConfig("test-pause-container")
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"sleep", "30"}

		err := client.CreateContainer(ctx, containerConfig)
		require.NoError(t, err)
		defer client.RemoveContainer(ctx, containerConfig, true)

		err = client.StartContainer(ctx, containerConfig)
		require.NoError(t, err)

		err = client.PauseContainer(ctx, containerConfig)
		require.NoError(t, err)

		err = client.UnpauseContainer(ctx, containerConfig)
		require.NoError(t, err)

		err = client.StopContainer(ctx, containerConfig)
		require.NoError(t, err)
	})
}

func TestNetworkOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)
	imageConfig := setupTestImage(ctx, t, client)

	networkConfig := network.NewConfig("test-network")

	// Create network
	err := client.CreateNetwork(ctx, networkConfig)
	require.NoError(t, err)
	require.NotEmpty(t, networkConfig.Id)
	defer client.RemoveNetwork(ctx, networkConfig)

	// Create a test container
	containerConfig := container.NewConfig("network-test-container")
	containerConfig.Options.Image = imageConfig.Ref
	containerConfig.Options.Cmd = []string{"sleep", "30"}

	err = client.CreateContainer(ctx, containerConfig)
	require.NoError(t, err)
	defer client.RemoveContainer(ctx, containerConfig, true)

	// Test network connect/disconnect
	err = client.ConnectContainerToNetwork(ctx, networkConfig, containerConfig)
	require.NoError(t, err)

	err = client.DisconnectContainerFromNetwork(ctx, networkConfig, containerConfig, true)
	require.NoError(t, err)
}

func TestVolumeOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)

	volumeConfig := volume.NewConfig("test-volume")

	// Create volume
	err := client.CreateVolume(ctx, volumeConfig)
	require.NoError(t, err)
	defer client.RemoveVolume(ctx, volumeConfig)

	// Test volume prune
	_, err = client.PruneVolumes(ctx, map[string][]string{
		"label": {"test=true"},
	})
	require.NoError(t, err)
}

func TestImageOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)

	t.Run("Image Pull and Tag", func(t *testing.T) {
		imageConfig, err := image.NewConfig("alpine:latest")
		require.NoError(t, err)

		err = client.PullImage(ctx, imageConfig)
		require.NoError(t, err)

		newTag := fmt.Sprintf("alpine:test-%d", time.Now().Unix())
		err = client.TagImage(ctx, imageConfig, newTag)
		require.NoError(t, err)

		taggedConfig, err := image.NewConfig(newTag)
		require.NoError(t, err)
		err = client.RemoveImage(ctx, taggedConfig, true)
		require.NoError(t, err)
	})

	t.Run("Image Save and Load", func(t *testing.T) {
		imageConfig, err := image.NewConfig("alpine:latest")
		require.NoError(t, err)

		tempFile := fmt.Sprintf("test-image-%d.tar", time.Now().Unix())
		defer os.Remove(tempFile)

		err = client.SaveImage(ctx, imageConfig, tempFile)
		require.NoError(t, err)

		err = client.LoadImage(ctx, tempFile)
		require.NoError(t, err)
	})
}

func TestCustomWriters(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)

	// Test custom writers
	var buf bytes.Buffer
	client.SetImageResponeWriter(&buf)
	client.SetStatsResponeWriter(&buf)
	client.SetLogResponseWriter(&buf)

	imageConfig, err := image.NewConfig("alpine:latest")
	require.NoError(t, err)

	err = client.PullImage(ctx, imageConfig)
	require.NoError(t, err)
	require.NotEmpty(t, buf.String())
}
