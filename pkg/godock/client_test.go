package godock

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aptd3v/godock/pkg/godock/container"
	"github.com/aptd3v/godock/pkg/godock/errdefs"
	"github.com/aptd3v/godock/pkg/godock/exec"
	"github.com/aptd3v/godock/pkg/godock/execoptions"
	"github.com/aptd3v/godock/pkg/godock/image"
	"github.com/aptd3v/godock/pkg/godock/network"
	"github.com/aptd3v/godock/pkg/godock/networkoptions"
	"github.com/aptd3v/godock/pkg/godock/networkoptions/endpointoptions"
	"github.com/aptd3v/godock/pkg/godock/volume"
	"github.com/aptd3v/godock/pkg/godock/volumeoptions"
	containerType "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
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
	imageConfig := image.NewConfig("alpine:latest")

	rc, err := client.ImagePull(ctx, imageConfig)
	require.NoError(t, err)
	defer rc.Close()
	_, err = io.Copy(os.Stdout, rc)
	if err != nil {
		t.Fatalf("failed to copy logs: %v", err)
	}

	return imageConfig
}

func TestClientCreation(t *testing.T) {
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		client, err := NewClient(ctx)
		if err != nil {
			t.Skipf("Docker daemon is not running: %v", err)
		}
		require.NotNil(t, client)
		require.NotNil(t, client.wrapped)
	})

	t.Run("Daemon Not Running", func(t *testing.T) {
		// Mock isDaemonRunning to simulate daemon not running
		origIsDaemonRunning := isDaemonRunning
		defer func() { isDaemonRunning = origIsDaemonRunning }()

		isDaemonRunning = func(ctx context.Context, client client.APIClient) (bool, error) {
			return false, fmt.Errorf("connection refused")
		}

		client, err := NewClient(ctx)
		require.Error(t, err)
		require.ErrorIs(t, err, errdefs.ErrDaemonNotRunning)

		var dne *errdefs.DaemonNotRunningError
		require.ErrorAs(t, err, &dne)
		require.Contains(t, dne.Error(), "docker daemon is not running: connection refused")
		require.Nil(t, client)
	})
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
		err := client.ContainerCreate(ctx, containerConfig)
		require.NoError(t, err)
		require.NotEmpty(t, containerConfig.Id)
		defer client.ContainerRemove(ctx, containerConfig, true)

		// Start container
		err = client.ContainerStart(ctx, containerConfig)
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

		err := client.ContainerCreate(ctx, containerConfig)
		require.NoError(t, err)
		defer client.ContainerRemove(ctx, containerConfig, true)

		err = client.ContainerStart(ctx, containerConfig)
		require.NoError(t, err)

		err = client.ContainerPause(ctx, containerConfig)
		require.NoError(t, err)

		err = client.ContainerUnpause(ctx, containerConfig)
		require.NoError(t, err)

		err = client.ContainerStop(ctx, containerConfig)
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
	err := client.NetworkCreate(ctx, networkConfig)
	require.NoError(t, err)
	require.NotEmpty(t, networkConfig.Id)
	defer client.NetworkRemove(ctx, networkConfig.Id)

	// Create a test container
	containerConfig := container.NewConfig("network-test-container")
	containerConfig.Options.Image = imageConfig.Ref
	containerConfig.Options.Cmd = []string{"sleep", "30"}

	err = client.ContainerCreate(ctx, containerConfig)
	require.NoError(t, err)
	defer client.ContainerRemove(ctx, containerConfig, true)

	err = client.ContainerStart(ctx, containerConfig)
	require.NoError(t, err)

	// Test network connect/disconnect
	err = client.NetworkConnectContainer(ctx, networkConfig.Id, containerConfig.Id, &endpointoptions.Endpoint{})
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

	err = client.NetworkDisconnectContainer(ctx, networkConfig.Id, containerConfig.Id, true)
	require.NoError(t, err)
}

func TestVolumeOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)

	t.Run("Basic Volume Operations", func(t *testing.T) {
		volumeConfig := volume.NewConfig("test-volume")
		err := client.VolumeCreate(ctx, volumeConfig)
		require.NoError(t, err)
		defer client.VolumeRemove(ctx, volumeConfig.Options.Name, true)
	})

	t.Run("Volume Prune with Label Filter", func(t *testing.T) {
		// Clean up any existing test volumes first
		existingVols, err := client.VolumeList(ctx)
		require.NoError(t, err)
		for _, vol := range existingVols.Volumes {
			if strings.HasPrefix(vol.Name, "test-") {
				t.Logf("Cleaning up existing volume: %s", vol.Name)
				_ = client.VolumeRemove(ctx, vol.Name, true)
			}
		}

		// Define test volume configurations
		volumes := []struct {
			name   string
			labels map[string]string
		}{
			{"test-keep", map[string]string{"keep": "true", "env": "prod"}},
			{"test-delete", map[string]string{"delete": "true", "env": "dev"}},
			{"test-env", map[string]string{"env": "stage"}},
			{"test-empty", map[string]string{}},
		}

		// Helper function to create test volumes
		createTestVolumes := func() {
			for _, v := range volumes {
				volumeConfig := volume.NewConfig(v.name)
				if len(v.labels) > 0 {
					volumeConfig.SetOptions(volumeoptions.SetLabels(v.labels))
				}
				err := client.VolumeCreate(ctx, volumeConfig)
				require.NoError(t, err)
				t.Logf("Created volume %s with labels %v", v.name, v.labels)

				// Verify volume was created with correct labels
				vol, err := client.wrapped.VolumeInspect(ctx, v.name)
				require.NoError(t, err)
				t.Logf("Inspected volume %s: Labels=%v", v.name, vol.Labels)
				if len(v.labels) > 0 {
					require.Equal(t, v.labels, vol.Labels)
				} else {
					require.Empty(t, vol.Labels)
				}
			}
		}

		// Test 1: Keep volumes with specific label and value
		t.Log("Test 1: Keep volumes with keep=true")
		createTestVolumes()
		report, err := client.VolumePrune(ctx, FilterIncludeLabelValue("keep", "true"))
		require.NoError(t, err)
		t.Logf("Pruned volumes: %v", report.VolumesDeleted)
		require.NotContains(t, report.VolumesDeleted, "test-keep", "Volume with keep=true should be kept")
		require.Contains(t, report.VolumesDeleted, "test-delete", "Volume without keep=true should be pruned")
		require.Contains(t, report.VolumesDeleted, "test-env", "Volume without keep=true should be pruned")
		require.Contains(t, report.VolumesDeleted, "test-empty", "Volume without keep=true should be pruned")

		// Test 2: Keep volumes with env label (any value)
		t.Log("Test 2: Keep volumes with env label")
		createTestVolumes()
		report, err = client.VolumePrune(ctx, FilterIncludeLabel("env"))
		require.NoError(t, err)
		t.Logf("Pruned volumes: %v", report.VolumesDeleted)
		require.NotContains(t, report.VolumesDeleted, "test-keep", "Volume with env label should be kept")
		require.NotContains(t, report.VolumesDeleted, "test-delete", "Volume with env label should be kept")
		require.NotContains(t, report.VolumesDeleted, "test-env", "Volume with env label should be kept")
		require.Contains(t, report.VolumesDeleted, "test-empty", "Volume without env label should be pruned")

		// Test 3: Keep volumes without keep label
		t.Log("Test 3: Keep volumes without keep label")
		createTestVolumes()
		report, err = client.VolumePrune(ctx, FilterExcludeLabel("keep"))
		require.NoError(t, err)
		t.Logf("Pruned volumes: %v", report.VolumesDeleted)
		require.Contains(t, report.VolumesDeleted, "test-keep", "Volume with keep label should be pruned")
		require.NotContains(t, report.VolumesDeleted, "test-delete", "Volume without keep label should be kept")
		require.NotContains(t, report.VolumesDeleted, "test-env", "Volume without keep label should be kept")
		require.NotContains(t, report.VolumesDeleted, "test-empty", "Volume without keep label should be kept")

		// Test 4: Keep volumes without env=prod
		t.Log("Test 4: Keep volumes without env=prod")
		createTestVolumes()
		report, err = client.VolumePrune(ctx, FilterExcludeLabelValue("env", "prod"))
		require.NoError(t, err)
		t.Logf("Pruned volumes: %v", report.VolumesDeleted)
		require.Contains(t, report.VolumesDeleted, "test-keep", "Volume with env=prod should be pruned")
		require.NotContains(t, report.VolumesDeleted, "test-delete", "Volume without env=prod should be kept")
		require.NotContains(t, report.VolumesDeleted, "test-env", "Volume without env=prod should be kept")
		require.NotContains(t, report.VolumesDeleted, "test-empty", "Volume without env=prod should be kept")

		// Test 5: Keep volumes with nonexistent label
		t.Log("Test 5: Keep volumes with nonexistent label")
		createTestVolumes()
		report, err = client.VolumePrune(ctx, FilterIncludeLabel("nonexistent"))
		require.NoError(t, err)
		t.Logf("Pruned volumes: %v", report.VolumesDeleted)
		require.Contains(t, report.VolumesDeleted, "test-keep", "Volume without nonexistent label should be pruned")
		require.Contains(t, report.VolumesDeleted, "test-delete", "Volume without nonexistent label should be pruned")
		require.Contains(t, report.VolumesDeleted, "test-env", "Volume without nonexistent label should be pruned")
		require.Contains(t, report.VolumesDeleted, "test-empty", "Volume without nonexistent label should be pruned")
	})
}

func TestImageOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)

	t.Run("Image Pull and Tag", func(t *testing.T) {
		imageConfig := image.NewConfig("alpine")

		rc, err := client.ImagePull(ctx, imageConfig)
		require.NoError(t, err)
		defer rc.Close()
		_, err = io.Copy(os.Stdout, rc)
		if err != nil {
			t.Fatalf("failed to copy logs: %v", err)
		}

		newTag := fmt.Sprintf("alpine:test-%d", time.Now().Unix())
		err = client.ImageTag(ctx, imageConfig, newTag)
		require.NoError(t, err)

		taggedConfig := image.NewConfig(newTag)
		require.NoError(t, err)
		_, err = client.ImageRemove(ctx, taggedConfig.Ref, true, true)
		require.NoError(t, err)
	})

	t.Run("Image Save and Load", func(t *testing.T) {
		imageConfig := image.NewConfig("alpine")

		tempFile := fmt.Sprintf("test-image-%d.tar", time.Now().Unix())
		defer os.Remove(tempFile)

		_, err := client.ImageSaveToReader(ctx, []string{imageConfig.Ref})
		require.NoError(t, err)

		_, err = client.ImageLoadFromReader(ctx, bytes.NewReader([]byte(tempFile)), true)
		require.NoError(t, err)
	})
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
		defer client.ContainerRemove(ctx, containerConfig, true) // Set up cleanup before any operations

		err := client.RunAndWait(ctx, containerConfig)
		require.NoError(t, err)
	})

	t.Run("RunAndWait with Timeout", func(t *testing.T) {
		containerConfig := container.NewConfig("test-run-and-wait-timeout")
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"sleep", "5"}
		defer client.ContainerRemove(ctx, containerConfig, true)

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
		defer client.ContainerRemove(ctx, containerConfig, true) // Set up cleanup before any operations

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

		err := client.ContainerCreate(ctx, containerConfig)
		require.NoError(t, err)
		defer client.ContainerRemove(ctx, containerConfig, true)
		err = client.ContainerStart(ctx, containerConfig)
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
			defer client.ContainerRemove(ctx, containerConfig, true) // Set up cleanup before any operations

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
			defer client.ContainerRemove(ctx, containerConfig, true) // Set up cleanup before any operations

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

	t.Run("GetImageSize", func(t *testing.T) {
		imageConfig := image.NewConfig("alpine")

		size, err := client.GetImageSize(ctx, imageConfig)
		require.NoError(t, err)
		require.Greater(t, size, int64(0))
	})

	t.Run("GetImageCreatedTime", func(t *testing.T) {
		imageConfig := image.NewConfig("alpine")

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
		err = client.NetworkCreate(ctx, networkConfig)
		require.NoError(t, err)
		defer client.NetworkRemove(ctx, networkConfig.Id)

		// Check after creation
		exists, err = client.IsNetworkExists(ctx, networkConfig)
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("GetNetworkContainers", func(t *testing.T) {
		networkConfig := network.NewConfig("test-network-containers")
		err := client.NetworkCreate(ctx, networkConfig)
		require.NoError(t, err)
		defer client.NetworkRemove(ctx, networkConfig.Id)
		t.Logf("Created network: %s (ID: %s)", networkConfig.Name, networkConfig.Id)

		// Create and connect a container
		containerConfig := container.NewConfig("test-network-container")
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"sleep", "1"}

		err = client.ContainerCreate(ctx, containerConfig)
		require.NoError(t, err)
		defer client.ContainerRemove(ctx, containerConfig, true)

		err = client.ContainerStart(ctx, containerConfig)
		require.NoError(t, err)

		err = client.NetworkConnectContainer(ctx, networkConfig.Id, containerConfig.Id, &endpointoptions.Endpoint{})
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
		err = client.VolumeCreate(ctx, volumeConfig)
		require.NoError(t, err)
		defer client.VolumeRemove(ctx, "test-volume-exists", true)

		// Check after creation
		exists, err = client.IsVolumeExists(ctx, volumeConfig)
		require.NoError(t, err)
		require.True(t, exists)
	})

	t.Run("GetVolumeUsage", func(t *testing.T) {
		volumeConfig := volume.NewConfig("test-volume-usage")
		volumeConfig.SetOptions(
			volumeoptions.SetDriver("local"),
			volumeoptions.AddLabel("test", "true"),
		)
		err := client.VolumeCreate(ctx, volumeConfig)
		require.NoError(t, err)
		defer client.VolumeRemove(ctx, "test-volume-usage", true)

		usage, err := client.VolumeUsage(ctx, "test-volume-usage")
		if err != nil && strings.Contains(err.Error(), "volume usage data not available") {
			t.Skip("volume usage data not available")
		}
		require.NoError(t, err)
		// Note: Size might be 0 if the driver doesn't support usage reporting
		require.GreaterOrEqual(t, usage.Size, int64(0))
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
			if err := client.ContainerStop(ctx, containerConfig); err != nil {
				t.Logf("Error stopping container: %v", err)
			}
			if err := client.ContainerRemove(ctx, containerConfig, true); err != nil {
				t.Logf("Error removing container: %v", err)
			}
		}()

		err := client.ContainerCreate(ctx, containerConfig)
		require.NoError(t, err)

		err = client.ContainerStart(ctx, containerConfig)
		require.NoError(t, err)

		// Wait for container to finish
		statusCh, errCh := client.ContainerWait(ctx, containerConfig)
		select {
		case err := <-errCh:
			require.NoError(t, err)
		case status := <-statusCh:
			require.Equal(t, int64(0), status.StatusCode)
		}

		// Get logs after container has finished
		logs, err := client.ContainerLogs(ctx, containerConfig)
		require.NoError(t, err)

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
			if err := client.ContainerStop(ctx, containerConfig); err != nil {
				t.Logf("Error stopping container: %v", err)
			}
			if err := client.ContainerRemove(ctx, containerConfig, true); err != nil {
				t.Logf("Error removing container: %v", err)
			}
		}()

		err := client.ContainerCreate(ctx, containerConfig)
		require.NoError(t, err)

		err = client.ContainerStart(ctx, containerConfig)
		require.NoError(t, err)
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
			if err := client.ContainerStop(ctx, containerConfig); err != nil {
				t.Logf("Error stopping container: %v", err)
			}
			if err := client.ContainerRemove(ctx, containerConfig, true); err != nil {
				t.Logf("Error removing container: %v", err)
			}
		}()

		err := client.ContainerCreate(ctx, containerConfig)
		require.NoError(t, err)

		err = client.ContainerStart(ctx, containerConfig)
		require.NoError(t, err)

		// Verify container is running
		running, err := client.IsContainerRunning(ctx, containerConfig)
		require.NoError(t, err)
		require.True(t, running)

		// Restart container
		err = client.ContainerRestart(ctx, containerConfig)
		require.NoError(t, err)

		// Give some time for the container to restart
		time.Sleep(100 * time.Millisecond)

		// Verify container is still running after restart
		running, err = client.IsContainerRunning(ctx, containerConfig)
		require.NoError(t, err)
		require.True(t, running)
	})
}

func TestContainerExecOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)
	imageConfig := setupTestImage(ctx, t, client)

	t.Run("ContainerExec Basic Operations", func(t *testing.T) {
		containerConfig := container.NewConfig("test-exec-container")
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"sleep", "30"}
		containerConfig.Options.Tty = true
		containerConfig.Options.OpenStdin = true
		containerConfig.Options.AttachStdin = true
		containerConfig.Options.AttachStdout = true
		containerConfig.Options.AttachStderr = true

		err := client.ContainerCreate(ctx, containerConfig)
		require.NoError(t, err)
		defer client.ContainerRemove(ctx, containerConfig, true)

		err = client.ContainerStart(ctx, containerConfig)
		require.NoError(t, err)

		// Wait for container to be fully running
		time.Sleep(2 * time.Second)

		// Create exec config
		execConfig := exec.NewConfig()
		execConfig.SetOptions(
			execoptions.CMD("echo", "hello"),
			execoptions.AttachStdout(true),
			execoptions.AttachStderr(true),
			execoptions.TTY(true),
		)

		// First create the exec instance
		execID, err := client.ContainerExecCreate(ctx, containerConfig, execConfig)
		require.NoError(t, err)
		require.NotEmpty(t, execID, "Exec ID should not be empty")
		require.Equal(t, execID, execConfig.ID, "Exec ID should match in config")
		t.Logf("Created exec instance with ID: %s", execID)

		// Then attach to it
		hijack, err := client.ContainerExecAttach(ctx, execID, execConfig)
		require.NoError(t, err)
		defer hijack.Close()

		// Start the exec instance
		err = client.ContainerExecStart(ctx, containerConfig, execConfig)
		require.NoError(t, err)

		// Read the output
		output, err := io.ReadAll(hijack.Reader)
		require.NoError(t, err)
		require.Contains(t, string(output), "hello")

		// Verify exec completed successfully
		inspect, err := client.ContainerExecInspect(ctx, execConfig)
		require.NoError(t, err)
		require.NotNil(t, inspect)
		require.Equal(t, 0, inspect.ExitCode)
	})
}

func TestContainerAdvancedOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)
	imageConfig := setupTestImage(ctx, t, client)

	t.Run("Container Update and Diff", func(t *testing.T) {
		containerConfig := container.NewConfig("test-advanced-container")
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"sh", "-c", "echo 'test' > /test.txt && sleep 5"}

		err := client.ContainerCreate(ctx, containerConfig)
		require.NoError(t, err)
		defer client.ContainerRemove(ctx, containerConfig, true)

		err = client.ContainerStart(ctx, containerConfig)
		require.NoError(t, err)

		// Test ContainerUpdate - set both memory and swap
		updateRes, err := client.ContainerUpdate(ctx, containerConfig, func(config *containerType.UpdateConfig) {
			config.Memory = 512 * 1024 * 1024      // 512MB
			config.MemorySwap = 1024 * 1024 * 1024 // 1GB
		})
		require.NoError(t, err)
		require.NotNil(t, updateRes)

		// Wait for container to create the file
		time.Sleep(2 * time.Second)

		// Test ContainerDiff
		diff, err := client.ContainerDiff(ctx, containerConfig)
		require.NoError(t, err)
		require.NotEmpty(t, diff)

		// Verify the file was created
		found := false
		for _, change := range diff {
			if change.Path == "/test.txt" {
				found = true
				break
			}
		}
		require.True(t, found, "Expected file /test.txt to be created")
	})

	t.Run("Container Kill and Rename", func(t *testing.T) {
		containerConfig := container.NewConfig("test-kill-container")
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"sleep", "30"}

		err := client.ContainerCreate(ctx, containerConfig)
		require.NoError(t, err)
		defer client.ContainerRemove(ctx, containerConfig, true)

		err = client.ContainerStart(ctx, containerConfig)
		require.NoError(t, err)

		// Wait for container to be fully running
		time.Sleep(2 * time.Second)

		// Verify container is running before rename
		running, err := client.IsContainerRunning(ctx, containerConfig)
		require.NoError(t, err)
		require.True(t, running, "Container should be running before rename")

		// Test ContainerRename
		newName := "test-renamed-container"
		err = client.ContainerRename(ctx, containerConfig, newName)
		require.NoError(t, err)

		// Verify rename
		inspect, err := client.ContainerInspect(ctx, containerConfig)
		require.NoError(t, err)
		require.Contains(t, inspect.Name, newName)

		// Test ContainerKill with SIGKILL to ensure immediate termination
		err = client.ContainerKill(ctx, containerConfig, "SIGKILL")
		require.NoError(t, err)

		// Wait for container to be killed and verify its state
		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			running, err = client.IsContainerRunning(ctx, containerConfig)
			require.NoError(t, err)
			if !running {
				break
			}
			t.Logf("Container still running, retry %d/%d", i+1, maxRetries)
			time.Sleep(time.Second)
		}

		require.False(t, running, "Container should be stopped after kill")

		// Double check with inspect
		inspect, err = client.ContainerInspect(ctx, containerConfig)
		require.NoError(t, err)
		require.False(t, inspect.State.Running, "Container state should be not running")
		t.Logf("Container final state: %+v", inspect.State)
	})

	t.Run("Container Top and Stats", func(t *testing.T) {
		containerConfig := container.NewConfig("test-top-container")
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"sleep", "30"}

		err := client.ContainerCreate(ctx, containerConfig)
		require.NoError(t, err)
		defer client.ContainerRemove(ctx, containerConfig, true)

		err = client.ContainerStart(ctx, containerConfig)
		require.NoError(t, err)

		// Wait for container to be fully running
		time.Sleep(2 * time.Second)

		// Test ContainerTop
		top, err := client.ContainerTop(ctx, containerConfig, []string{})
		require.NoError(t, err)
		require.NotNil(t, top)
		require.NotEmpty(t, top.Titles)

		// Test ContainerStatsOneShot
		stats, err := client.ContainerStatsOneShot(ctx, containerConfig)
		require.NoError(t, err)
		require.NotEmpty(t, stats)

		// Test ContainerStatsChan with timeout context
		ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		statsCh, errCh := client.ContainerStatsChan(ctxTimeout, containerConfig)
		require.NotNil(t, statsCh)
		require.NotNil(t, errCh)

		// Read first stat
		select {
		case stat := <-statsCh:
			require.NotEmpty(t, stat)
		case err := <-errCh:
			require.NoError(t, err)
		case <-ctxTimeout.Done():
			t.Fatal("Timeout waiting for stats")
		}
	})
}

func TestImageAdvancedOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)
	imageConfig := setupTestImage(ctx, t, client)

	t.Run("Image History and Inspect", func(t *testing.T) {
		// Test ImageHistory
		history, err := client.ImageHistory(ctx, imageConfig.Ref)
		require.NoError(t, err)
		require.NotEmpty(t, history)

		// Test ImageInspect
		inspect, err := client.ImageInspect(ctx, imageConfig.Ref)
		require.NoError(t, err)
		require.NotNil(t, inspect)
		require.NotEmpty(t, inspect.ID)
	})

	t.Run("Image Commit", func(t *testing.T) {
		containerConfig := container.NewConfig("test-commit-container")
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"touch", "/test-file"}

		err := client.ContainerCreate(ctx, containerConfig)
		require.NoError(t, err)
		defer client.ContainerRemove(ctx, containerConfig, true)

		err = client.ContainerStart(ctx, containerConfig)
		require.NoError(t, err)

		// Wait for container to create the file
		time.Sleep(2 * time.Second)

		// Stop the container before committing
		err = client.ContainerStop(ctx, containerConfig)
		require.NoError(t, err)

		// Test ImageCommit
		newImageConfig := image.NewConfig("test-committed-image:latest")
		imageID, err := client.ImageCommit(ctx, containerConfig, newImageConfig, func(opts *containerType.CommitOptions) {
			opts.Reference = "test-committed-image:latest"
		})
		require.NoError(t, err)
		require.NotEmpty(t, imageID)

		// Verify the committed image exists
		inspect, err := client.ImageInspect(ctx, newImageConfig.Ref)
		require.NoError(t, err)
		require.Equal(t, imageID, inspect.ID)

		// Now cleanup the committed image
		_, err = client.ImageRemove(ctx, newImageConfig.Ref, true, true)
		require.NoError(t, err)
	})

	t.Run("Image Prune", func(t *testing.T) {
		// Test ImagesPrune
		report, err := client.ImagesPrune(ctx)
		require.NoError(t, err)
		require.NotNil(t, report)
	})
}

func TestListOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)
	imageConfig := setupTestImage(ctx, t, client)

	t.Run("Container List Operations", func(t *testing.T) {
		// Create a test container
		containerConfig := container.NewConfig("test-list-container")
		containerConfig.Options.Image = imageConfig.Ref
		containerConfig.Options.Cmd = []string{"sleep", "30"}
		containerConfig.Options.Labels = map[string]string{"test": "true"}

		err := client.ContainerCreate(ctx, containerConfig)
		require.NoError(t, err)
		t.Logf("Created container with ID: %s", containerConfig.Id)
		defer client.ContainerRemove(ctx, containerConfig, true)

		err = client.ContainerStart(ctx, containerConfig)
		require.NoError(t, err)
		t.Log("Started container")

		// Wait for container to be fully running
		time.Sleep(2 * time.Second)

		// Verify container is running and has correct labels
		inspect, err := client.ContainerInspect(ctx, containerConfig)
		require.NoError(t, err)
		require.True(t, inspect.State.Running, "Container should be running")
		require.Equal(t, map[string]string{"test": "true"}, inspect.Config.Labels, "Container should have correct labels")
		t.Logf("Container state: %s, Labels: %v", inspect.State.Status, inspect.Config.Labels)

		// Test basic container list
		containers, err := client.ContainerList(ctx, WithContainerAll(true))
		require.NoError(t, err)
		require.NotEmpty(t, containers)
		t.Logf("Found %d containers in basic list", len(containers))

		// Test with label filter
		containers, err = client.ContainerList(ctx,
			WithContainerAll(true),
			WithContainerFilter("label", "test=true"))
		require.NoError(t, err)
		t.Logf("Found %d containers with label filter", len(containers))
		require.NotEmpty(t, containers, "No containers found with label filter")

		// Verify our container is in the list
		var found bool
		for _, c := range containers {
			t.Logf("Container in filtered list: ID=%s, Labels=%v", c.ID, c.Labels)
			if c.ID == containerConfig.Id {
				found = true
				require.Equal(t, map[string]string{"test": "true"}, c.Labels)
				break
			}
		}
		require.True(t, found, "Test container not found in list")

		// Stop container before pruning
		err = client.ContainerStop(ctx, containerConfig)
		require.NoError(t, err)
		t.Log("Stopped container")

		// Wait for container to stop
		time.Sleep(2 * time.Second)

		// Verify container is stopped
		inspect, err = client.ContainerInspect(ctx, containerConfig)
		require.NoError(t, err)
		require.False(t, inspect.State.Running, "Container should be stopped")
		t.Logf("Container state after stop: %s", inspect.State.Status)

		// Test ContainerPrune
		report, err := client.ContainerPrune(ctx, WithPruneFilter("label", "test=true"))
		require.NoError(t, err)
		require.NotNil(t, report)
		t.Logf("Prune report: deleted=%v, space reclaimed=%d", report.ContainersDeleted, report.SpaceReclaimed)
		require.Contains(t, report.ContainersDeleted, containerConfig.Id, "Container should have been pruned")
	})

	t.Run("Network List Operations", func(t *testing.T) {
		// Create a test network
		networkConfig := network.NewConfig("test-list-network")
		err := client.NetworkCreate(ctx, networkConfig)
		require.NoError(t, err)
		defer client.NetworkRemove(ctx, networkConfig.Id)

		// Test NetworkList with various options
		networks, err := client.NetworkList(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, networks)

		// Test NetworkList with filter
		networks, err = client.NetworkList(ctx, WithNetworkFilter("name", networkConfig.Name))
		require.NoError(t, err)
		require.NotEmpty(t, networks)
		found := false
		for _, n := range networks {
			if n.ID == networkConfig.Id {
				found = true
				break
			}
		}
		require.True(t, found, "Network not found in list")

		// Test NetworkInspect
		inspect, err := client.NetworkInspect(ctx, networkConfig.Id, WithNetworkInspectVerbose())
		require.NoError(t, err)
		require.Equal(t, networkConfig.Id, inspect.ID)
	})

	t.Run("Volume List Operations", func(t *testing.T) {
		// Create a test volume
		volumeConfig := volume.NewConfig("test-list-volume")
		volumeConfig.SetOptions(volumeoptions.AddLabel("test", "true"))
		err := client.VolumeCreate(ctx, volumeConfig)
		require.NoError(t, err)
		defer client.VolumeRemove(ctx, volumeConfig.Options.Name, true)

		// Test VolumeList with various options
		volumes, err := client.VolumeList(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, volumes.Volumes)

		// Test VolumeList with filter
		volumes, err = client.VolumeList(ctx, WithVolumeFilter("label", "test=true"))
		require.NoError(t, err)
		require.NotEmpty(t, volumes.Volumes)
		found := false
		for _, v := range volumes.Volumes {
			if v.Name == volumeConfig.Options.Name {
				found = true
				break
			}
		}
		require.True(t, found, "Volume not found in list")
	})

	t.Run("Image List Operations", func(t *testing.T) {
		// Test ImageList with various options
		images, err := client.ImageList(ctx, WithImageAll(true))
		require.NoError(t, err)
		require.NotEmpty(t, images)

		// Log all available images and their tags
		t.Log("Available images:")
		for _, img := range images {
			t.Logf("Image ID: %s, RepoTags: %v", img.ID, img.RepoTags)
		}
		t.Logf("Looking for image reference: %s", imageConfig.Ref)

		// Verify our test image is in the list
		var found bool
		for _, img := range images {
			for _, tag := range img.RepoTags {
				if tag == imageConfig.Ref {
					found = true
					t.Logf("Found matching image with tag: %s", tag)
					break
				}
			}
			if found {
				break
			}
		}
		require.True(t, found, "Image not found in list")

		// Test with filter
		images, err = client.ImageList(ctx,
			WithImageAll(true),
			WithImageFilter("reference", imageConfig.Ref))
		require.NoError(t, err)
		require.NotEmpty(t, images, "No images found with filter")

		// Verify filtered list contains our image
		found = false
		for _, img := range images {
			for _, tag := range img.RepoTags {
				if tag == imageConfig.Ref {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		require.True(t, found, "Image not found in filtered list")
	})
}

func TestImageSearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	client := setupTestClient(t)

	t.Run("Basic Search", func(t *testing.T) {
		results, err := client.ImageSearch(ctx, "alpine")
		require.NoError(t, err)
		require.NotEmpty(t, results)

		// Verify we got alpine in results
		var found bool
		for _, result := range results {
			if result.Name == "alpine" {
				found = true
				break
			}
		}
		require.True(t, found, "Expected to find official alpine image in search results")
	})

	t.Run("Search with Limit", func(t *testing.T) {
		limit := 5
		results, err := client.ImageSearch(ctx, "nginx", WithSearchLimit(limit))
		require.NoError(t, err)
		require.LessOrEqual(t, len(results), limit)
	})

	t.Run("Search with Filters", func(t *testing.T) {
		filters := filters.NewArgs()
		filters.Add("is-official", "true")

		results, err := client.ImageSearch(ctx, "ubuntu", WithSearchFilters(filters))
		require.NoError(t, err)
		require.NotEmpty(t, results)

		// Verify all results are official
		for _, result := range results {
			require.True(t, result.IsOfficial, "Expected only official images in results")
		}
	})

	t.Run("Search Non-Existent Image", func(t *testing.T) {
		results, err := client.ImageSearch(ctx, "nonexistentimageqwerty123456")
		require.NoError(t, err)
		require.Empty(t, results)
	})
}
