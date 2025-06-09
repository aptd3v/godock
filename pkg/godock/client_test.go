package godock

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/aptd3v/godock/pkg/godock/container"
	"github.com/aptd3v/godock/pkg/godock/image"
	"github.com/aptd3v/godock/pkg/godock/network"
	"github.com/aptd3v/godock/pkg/godock/networkoptions"
	"github.com/aptd3v/godock/pkg/godock/volume"
	"github.com/google/uuid"
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
	networkConfig.SetOptions(
		networkoptions.Driver("bridge"),
		networkoptions.Attachable(),
	)

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

	err = client.StartContainer(ctx, containerConfig)
	require.NoError(t, err)

	// Test network connect/disconnect
	err = client.ConnectContainerToNetwork(ctx, networkConfig, containerConfig)
	require.NoError(t, err)
	t.Logf("Connected container %s to network %s", containerConfig.Id, networkConfig.Id)

	// Retry a few times with delay to ensure network is updated
	var containers []string
	var retryErr error
	for i := 0; i < 5; i++ {
		containers, retryErr = client.GetNetworkContainers(ctx, networkConfig)
		t.Logf("Attempt %d: Found %d containers. Error: %v", i+1, len(containers), retryErr)
		if len(containers) > 0 {
			t.Logf("Container IDs found: %v", containers)
		}
		if retryErr == nil && len(containers) > 0 && containers[0] != "" {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	require.NoError(t, retryErr)
	require.NotEmpty(t, containers, "No containers found in network after 5 retries")
	require.Contains(t, containers, containerConfig.Id)

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

func TestContainerUtilities(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)
	imageConfig := setupTestImage(ctx, t, client)

	t.Run("RunAndWait Success", func(t *testing.T) {
		containerConfig := container.NewConfig("test-run-and-wait")
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"echo", "hello"}
		defer client.RemoveContainer(ctx, containerConfig, true) // Set up cleanup before any operations

		err := client.RunAndWait(ctx, containerConfig)
		require.NoError(t, err)
	})

	t.Run("RunAndWait with Timeout", func(t *testing.T) {
		containerConfig := container.NewConfig("test-run-and-wait-timeout")
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"sleep", "5"}
		defer client.RemoveContainer(ctx, containerConfig, true)

		timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()

		err := client.RunAndWait(timeoutCtx, containerConfig)
		require.Error(t, err)
		require.Contains(t, err.Error(), context.DeadlineExceeded.Error())
	})

	t.Run("RunAsync Success", func(t *testing.T) {
		containerConfig := container.NewConfig("test-run-async")
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"echo", "hello"}
		defer client.RemoveContainer(ctx, containerConfig, true) // Set up cleanup before any operations

		resultCh, err := client.RunAsync(ctx, containerConfig)
		require.NoError(t, err)

		// Wait for result
		err = <-resultCh
		require.NoError(t, err)
	})

	t.Run("IsContainerRunning", func(t *testing.T) {
		containerConfig := container.NewConfig("test-container-running")
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"sleep", "2"}

		err := client.CreateContainer(ctx, containerConfig)
		require.NoError(t, err)
		defer client.RemoveContainer(ctx, containerConfig, true)

		err = client.StartContainer(ctx, containerConfig)
		require.NoError(t, err)

		// Check running state
		running, err := client.IsContainerRunning(ctx, containerConfig)
		require.NoError(t, err)
		require.True(t, running)

		// Wait for container to finish
		time.Sleep(3 * time.Second)

		// Check stopped state
		running, err = client.IsContainerRunning(ctx, containerConfig)
		require.NoError(t, err)
		require.False(t, running)
	})

	t.Run("GetContainerExitCode", func(t *testing.T) {
		t.Run("Success Exit", func(t *testing.T) {
			containerConfig := container.NewConfig("test-exit-success")
			containerConfig.Options.Image = imageConfig.Ref
			containerConfig.Options.Cmd = []string{"echo", "hello"}
			defer client.RemoveContainer(ctx, containerConfig, true) // Set up cleanup before any operations

			err := client.RunAndWait(ctx, containerConfig)
			require.NoError(t, err)

			exitCode, err := client.GetContainerExitCode(ctx, containerConfig)
			require.NoError(t, err)
			require.Equal(t, 0, exitCode)
		})

		t.Run("Error Exit", func(t *testing.T) {
			containerConfig := container.NewConfig("test-exit-error")
			containerConfig.Options.Image = imageConfig.Ref
			containerConfig.Options.Cmd = []string{"sh", "-c", "exit 1"}
			defer client.RemoveContainer(ctx, containerConfig, true) // Set up cleanup before any operations

			err := client.RunAndWait(ctx, containerConfig)
			require.Error(t, err)

			exitCode, err := client.GetContainerExitCode(ctx, containerConfig)
			require.NoError(t, err)
			require.Equal(t, 1, exitCode)
		})
	})
}

func TestImageUtilities(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)

	t.Run("PullImageIfNotPresent", func(t *testing.T) {
		imageConfig, err := image.NewConfig("alpine:latest")
		require.NoError(t, err)

		// First pull should download
		err = client.PullImageIfNotPresent(ctx, imageConfig)
		require.NoError(t, err)

		// Second pull should be no-op
		err = client.PullImageIfNotPresent(ctx, imageConfig)
		require.NoError(t, err)
	})

	t.Run("GetImageSize", func(t *testing.T) {
		imageConfig, err := image.NewConfig("alpine:latest")
		require.NoError(t, err)

		size, err := client.GetImageSize(ctx, imageConfig)
		require.NoError(t, err)
		require.Greater(t, size, int64(0))
	})

	t.Run("GetImageCreatedTime", func(t *testing.T) {
		imageConfig, err := image.NewConfig("alpine:latest")
		require.NoError(t, err)

		created, err := client.GetImageCreatedTime(ctx, imageConfig)
		require.NoError(t, err)
		require.NotEmpty(t, created)
	})
}

func TestNetworkUtilities(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)
	imageConfig := setupTestImage(ctx, t, client)

	t.Run("IsNetworkExists", func(t *testing.T) {
		networkConfig := network.NewConfig("test-network-exists")

		// Check before creation
		exists, err := client.IsNetworkExists(ctx, networkConfig)
		require.NoError(t, err)
		require.False(t, exists)

		// Create network
		err = client.CreateNetwork(ctx, networkConfig)
		require.NoError(t, err)
		defer client.RemoveNetwork(ctx, networkConfig)

		// Check after creation
		exists, err = client.IsNetworkExists(ctx, networkConfig)
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("GetNetworkContainers", func(t *testing.T) {
		networkConfig := network.NewConfig("test-network-containers")
		err := client.CreateNetwork(ctx, networkConfig)
		require.NoError(t, err)
		defer client.RemoveNetwork(ctx, networkConfig)
		t.Logf("Created network: %s (ID: %s)", networkConfig.Name, networkConfig.Id)

		// Create and connect a container
		containerConfig := container.NewConfig("test-network-container")
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"sleep", "1"}

		err = client.CreateContainer(ctx, containerConfig)
		require.NoError(t, err)
		defer client.RemoveContainer(ctx, containerConfig, true)

		err = client.StartContainer(ctx, containerConfig)
		require.NoError(t, err)

		err = client.ConnectContainerToNetwork(ctx, networkConfig, containerConfig)
		require.NoError(t, err)

		containers, err := client.GetNetworkContainers(ctx, networkConfig)
		require.NoError(t, err)
		require.Contains(t, containers, containerConfig.Id)
	})
}

func TestVolumeUtilities(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)

	t.Run("IsVolumeExists", func(t *testing.T) {
		volumeConfig := volume.NewConfig("test-volume-exists")

		// Check before creation
		exists, err := client.IsVolumeExists(ctx, volumeConfig)
		require.NoError(t, err)
		require.False(t, exists)

		// Create volume
		err = client.CreateVolume(ctx, volumeConfig)
		require.NoError(t, err)
		defer client.RemoveVolume(ctx, volumeConfig)

		// Check after creation
		exists, err = client.IsVolumeExists(ctx, volumeConfig)
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("GetVolumeUsage", func(t *testing.T) {
		volumeConfig := volume.NewConfig("test-volume-usage")

		err := client.CreateVolume(ctx, volumeConfig)
		require.NoError(t, err)
		defer client.RemoveVolume(ctx, volumeConfig)

		size, err := client.GetVolumeUsage(ctx, volumeConfig)
		require.NoError(t, err)
		// Note: Size might be 0 if the driver doesn't support usage reporting
		require.GreaterOrEqual(t, size, int64(0))
	})
}

func TestContainerLogsAndStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)
	imageConfig := setupTestImage(ctx, t, client)

	t.Run("GetContainerLogs", func(t *testing.T) {
		containerConfig := container.NewConfig("test-logs-" + uuid.New().String())
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"sh", "-c", "echo stdout message && echo stderr message >&2"}
		containerConfig.Options.AttachStdout = true
		containerConfig.Options.AttachStderr = true
		containerConfig.Options.Tty = false

		t.Logf("Created container config: name=%s, image=%s, cmd=%v", containerConfig.Name, containerConfig.Options.Image, containerConfig.Options.Cmd)

		defer func() {
			t.Logf("Cleaning up container: %s", containerConfig.Id)
			if err := client.StopContainer(ctx, containerConfig); err != nil {
				t.Logf("Error stopping container: %v", err)
			}
			if err := client.RemoveContainer(ctx, containerConfig, true); err != nil {
				t.Logf("Error removing container: %v", err)
			}
		}()

		err := client.CreateContainer(ctx, containerConfig)
		require.NoError(t, err)

		err = client.StartContainer(ctx, containerConfig)
		require.NoError(t, err)

		// Wait for container to finish
		statusCh, errCh := client.ContainerWait(ctx, containerConfig)
		select {
		case err := <-errCh:
			require.NoError(t, err)
		case status := <-statusCh:
			require.Equal(t, int64(0), status.StatusCode)
		}

		// Test with custom writer
		var buf bytes.Buffer
		client.SetLogResponseWriter(&buf)

		// Get logs after container has finished
		logs, err := client.GetContainerLogs(ctx, containerConfig)
		require.NoError(t, err)
		defer logs.Close()

		// Read all logs
		logBytes, err := io.ReadAll(logs)
		require.NoError(t, err)
		logContent := string(logBytes)

		t.Logf("Log content: %s", logContent)
		require.Contains(t, logContent, "stdout message")
		require.Contains(t, logContent, "stderr message")
	})

	t.Run("GetContainerStats", func(t *testing.T) {
		containerConfig := container.NewConfig("test-stats-" + uuid.New().String())
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"sh", "-c", "while true; do echo 'consuming cpu'; done"}
		t.Logf("Created container config: name=%s, image=%s, cmd=%v", containerConfig.Name, containerConfig.Options.Image, containerConfig.Options.Cmd)

		defer func() {
			t.Logf("Cleaning up container: %s", containerConfig.Id)
			if err := client.StopContainer(ctx, containerConfig); err != nil {
				t.Logf("Error stopping container: %v", err)
			}
			if err := client.RemoveContainer(ctx, containerConfig, true); err != nil {
				t.Logf("Error removing container: %v", err)
			}
		}()

		err := client.CreateContainer(ctx, containerConfig)
		require.NoError(t, err)

		err = client.StartContainer(ctx, containerConfig)
		require.NoError(t, err)

		// Test with custom writer
		var buf bytes.Buffer
		client.SetStatsResponeWriter(&buf)

		// Create a context with timeout for stats collection
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		err = client.GetContainerStats(ctxWithTimeout, containerConfig)
		require.Error(t, err) // Should error due to context timeout
		require.NotEmpty(t, buf.String())
	})
}

func TestContainerRestartAndLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)
	imageConfig := setupTestImage(ctx, t, client)

	t.Run("RestartContainer", func(t *testing.T) {
		containerConfig := container.NewConfig("test-restart-" + uuid.New().String())
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"sleep", "10"}
		t.Logf("Created container config: name=%s, image=%s, cmd=%v", containerConfig.Name, containerConfig.Options.Image, containerConfig.Options.Cmd)

		defer func() {
			t.Logf("Cleaning up container: %s", containerConfig.Id)
			if err := client.StopContainer(ctx, containerConfig); err != nil {
				t.Logf("Error stopping container: %v", err)
			}
			if err := client.RemoveContainer(ctx, containerConfig, true); err != nil {
				t.Logf("Error removing container: %v", err)
			}
		}()

		err := client.CreateContainer(ctx, containerConfig)
		require.NoError(t, err)

		err = client.StartContainer(ctx, containerConfig)
		require.NoError(t, err)

		// Verify container is running
		running, err := client.IsContainerRunning(ctx, containerConfig)
		require.NoError(t, err)
		require.True(t, running)

		// Restart container
		err = client.RestartContainer(ctx, containerConfig)
		require.NoError(t, err)

		// Give some time for the container to restart
		time.Sleep(100 * time.Millisecond)

		// Verify container is still running after restart
		running, err = client.IsContainerRunning(ctx, containerConfig)
		require.NoError(t, err)
		require.True(t, running)
	})
}
