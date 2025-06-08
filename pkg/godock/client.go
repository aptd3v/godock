package godock

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aptd3v/godock/pkg/godock/container"
	"github.com/aptd3v/godock/pkg/godock/image"
	"github.com/aptd3v/godock/pkg/godock/network"
	"github.com/aptd3v/godock/pkg/godock/volume"
	containerType "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	imageType "github.com/docker/docker/api/types/image"
	dockerNetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// ImageProgress represents the JSON output from Docker image operations
type ImageProgress struct {
	Stream   string `json:"stream,omitempty"`
	Status   string `json:"status,omitempty"`
	Progress string `json:"progress,omitempty"`
	Aux      struct {
		ID string `json:"id,omitempty"`
	} `json:"aux,omitempty"`
	ErrorDetail struct {
		Message string `json:"message,omitempty"`
	} `json:"errorDetail,omitempty"`
	Error string `json:"error,omitempty"`
}

type Client struct {
	wrapped        *client.Client
	imageResWriter io.Writer
	statsResWriter io.Writer
	logResWriter   io.Writer
}

func (c *Client) CreateContainer(ctx context.Context, containerConfig *container.ContainerConfig) error {
	res, err := c.wrapped.ContainerCreate(
		ctx,
		containerConfig.Options,
		containerConfig.HostOptions,
		containerConfig.NetworkingOptions,
		containerConfig.PlatformOptions,
		containerConfig.Name,
	)
	if err != nil {
		return err
	}

	containerConfig.Id = res.ID

	return nil
}
func (c *Client) StartContainer(ctx context.Context, containerConfig *container.ContainerConfig) error {
	return c.wrapped.ContainerStart(ctx, containerConfig.Id, containerType.StartOptions{})
}

// GetContainerStats gets stats and is synchronus
// This is a blocking call and will return when the container is stopped or the context is cancelled
func (c *Client) GetContainerStats(ctx context.Context, containerConfig *container.ContainerConfig) error {
	res, err := c.wrapped.ContainerStats(ctx, containerConfig.Id, true)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if _, err := io.Copy(c.statsResWriter, res.Body); err != nil {
		return err
	}
	return nil
}

// GetContainerLogs returns a ReadCloser for container logs. If a custom log writer is configured,
// logs will also be written to it asynchronously. Caller is responsible for closing the returned reader.
func (c *Client) GetContainerLogs(ctx context.Context, containerConfig *container.ContainerConfig) (io.ReadCloser, error) {
	rc, err := c.wrapped.ContainerLogs(ctx, containerConfig.Id, containerType.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return nil, err
	}

	// Create a pipe to tee the output
	pr, pw := io.Pipe()

	// Start copying in background
	go func() {
		defer func() {
			rc.Close()
			pw.Close()
		}()

		_, err := io.Copy(io.MultiWriter(pw, c.logResWriter), rc)
		if err != nil && err != io.ErrClosedPipe {
			fmt.Printf("Error copying container logs: %v\n", err)
		}
	}()

	return pr, nil
}
func (c *Client) RemoveContainer(ctx context.Context, containerConfig *container.ContainerConfig, force bool) error {
	return c.wrapped.ContainerRemove(ctx, containerConfig.Id, containerType.RemoveOptions{
		RemoveVolumes: force,
		Force:         force,
	})
}
func (c *Client) UnpauseContainer(ctx context.Context, containerConfig *container.ContainerConfig) error {
	return c.wrapped.ContainerUnpause(ctx, containerConfig.Id)
}
func (c *Client) PauseContainer(ctx context.Context, containerConfig *container.ContainerConfig) error {
	return c.wrapped.ContainerPause(ctx, containerConfig.Id)
}
func (c *Client) RestartContainer(ctx context.Context, containerConfig *container.ContainerConfig) error {
	return c.wrapped.ContainerRestart(ctx, containerConfig.Id, containerType.StopOptions{})
}

func (c *Client) StopContainer(ctx context.Context, containerConfig *container.ContainerConfig) error {
	return c.wrapped.ContainerStop(ctx, containerConfig.Id, containerType.StopOptions{})
}

// ContainerWait waits for a container to finish and returns a channel for status and errors
func (c *Client) ContainerWait(ctx context.Context, containerConfig *container.ContainerConfig) (<-chan containerType.WaitResponse, <-chan error) {
	return c.wrapped.ContainerWait(ctx, containerConfig.Id, containerType.WaitConditionNotRunning)
}

func (c *Client) CreateNetwork(ctx context.Context, networkConfig *network.NetworkConfig) error {
	res, err := c.wrapped.NetworkCreate(ctx, networkConfig.Name, *networkConfig.Options)
	if err != nil {
		return err
	}
	networkConfig.Id = res.ID
	return nil
}

func (c *Client) CreateVolume(ctx context.Context, volumeConfig *volume.VolumeConfig) error {
	_, err := c.wrapped.VolumeCreate(ctx, *volumeConfig.Options)
	if err != nil {
		return err
	}
	return nil
}

// SetImageResponeWriter sets the image response writer for Docker's API.
// If this is not set, the client wrapper will default to stdout.
func (c *Client) SetImageResponeWriter(dst io.Writer) {
	c.imageResWriter = dst
}

// This sets the stats response writer for Docker's API.
// If this is not set, the client wrapper will default to StatsFormatter.
func (c *Client) SetStatsResponeWriter(dst io.Writer) {
	c.statsResWriter = dst
}

// This sets the log output response writer for Docker's API.
// If this is not set, the client wrapper will default to stdout.
func (c *Client) SetLogResponseWriter(dst io.Writer) {
	c.logResWriter = stdcopy.NewStdWriter(dst, stdcopy.Stdout)
}

func (c *Client) PullImage(ctx context.Context, imageConfig *image.ImageConfig) error {
	rc, err := c.wrapped.ImagePull(ctx, imageConfig.Ref, *imageConfig.PullOptions)
	if err != nil {
		return err
	}
	defer rc.Close()

	decoder := json.NewDecoder(rc)
	for {
		var progress ImageProgress
		if err := decoder.Decode(&progress); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if progress.Error != "" {
			return fmt.Errorf("pull error: %s", progress.Error)
		}
		if progress.Status != "" {
			fmt.Fprintf(c.imageResWriter, "%s\n", progress.Status)
		}
	}
	return nil
}
func (c *Client) BuildImage(ctx context.Context, imageConfig *image.ImageConfig) error {
	if imageConfig.BuildOptions.Context == nil {
		return errors.New("no build context was supplied use image.NewImageFromSrc(dir) or supply the context manually")
	}
	res, err := c.wrapped.ImageBuild(ctx, imageConfig.BuildOptions.Context, *imageConfig.BuildOptions)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if _, err = io.Copy(c.imageResWriter, res.Body); err != nil {
		return err
	}
	return nil
}

func (c *Client) String() string {
	return c.wrapped.DaemonHost()
}

func NewClient(ctx context.Context) (*Client, error) {
	c, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating new docker client %s", err)
	}
	ok, err := isDaemonRunning(ctx, c)
	if !ok {
		return nil, err
	}
	return &Client{
		wrapped:        c,
		imageResWriter: os.Stdout,
		statsResWriter: StatsFormatter(os.Stdout),
		logResWriter:   stdcopy.NewStdWriter(os.Stdout, stdcopy.Stdout),
	}, nil

}

// Unwraps the abstracted client for use with other docker packages
func (c *Client) Unwrap() client.APIClient {
	return c.wrapped
}

// checks if the docker daemon is running by pinging it
func isDaemonRunning(ctx context.Context, client *client.Client) (bool, error) {
	if _, err := client.Ping(ctx); err != nil {
		return false, fmt.Errorf("daemon is not running %s", err)
	}
	return true, nil
}

// Network Operations

func (c *Client) RemoveNetwork(ctx context.Context, networkConfig *network.NetworkConfig) error {
	return c.wrapped.NetworkRemove(ctx, networkConfig.Id)
}

func (c *Client) ConnectContainerToNetwork(ctx context.Context, networkConfig *network.NetworkConfig, containerConfig *container.ContainerConfig) error {
	// Create endpoint settings
	endpointSettings := &dockerNetwork.EndpointSettings{
		NetworkID: networkConfig.Id,
	}

	err := c.wrapped.NetworkConnect(ctx, networkConfig.Id, containerConfig.Id, endpointSettings)
	if err != nil {
		return fmt.Errorf("failed to connect container to network: %w", err)
	}

	// Verify connection
	network, err := c.wrapped.NetworkInspect(ctx, networkConfig.Id, dockerNetwork.InspectOptions{})
	if err != nil {
		return fmt.Errorf("failed to verify network connection: %w", err)
	}

	// Verify the container is in the network
	if _, exists := network.Containers[containerConfig.Id]; !exists {
		return fmt.Errorf("container %s not found in network %s after connection", containerConfig.Id, networkConfig.Id)
	}

	return nil
}

func (c *Client) DisconnectContainerFromNetwork(ctx context.Context, networkConfig *network.NetworkConfig, containerConfig *container.ContainerConfig, force bool) error {
	return c.wrapped.NetworkDisconnect(ctx, networkConfig.Id, containerConfig.Id, force)
}

// Volume Operations

func (c *Client) RemoveVolume(ctx context.Context, volumeConfig *volume.VolumeConfig) error {
	return c.wrapped.VolumeRemove(ctx, volumeConfig.Options.Name, false)
}

func (c *Client) PruneVolumes(ctx context.Context, filterMap map[string][]string) (uint64, error) {
	args := filters.NewArgs()
	for k, v := range filterMap {
		for _, val := range v {
			args.Add(k, val)
		}
	}
	report, err := c.wrapped.VolumesPrune(ctx, args)
	if err != nil {
		return 0, err
	}
	return report.SpaceReclaimed, nil
}

// Image Operations

func (c *Client) PushImage(ctx context.Context, imageConfig *image.ImageConfig) error {
	rc, err := c.wrapped.ImagePush(ctx, imageConfig.Ref, *imageConfig.PushOptions)
	if err != nil {
		return err
	}
	defer rc.Close()
	if _, err = io.Copy(c.imageResWriter, rc); err != nil {
		return err
	}
	return nil
}

func (c *Client) RemoveImage(ctx context.Context, imageConfig *image.ImageConfig, force bool) error {
	options := imageType.RemoveOptions{
		Force:         force,
		PruneChildren: true,
	}
	_, err := c.wrapped.ImageRemove(ctx, imageConfig.Ref, options)
	return err
}

func (c *Client) TagImage(ctx context.Context, imageConfig *image.ImageConfig, newTag string) error {
	return c.wrapped.ImageTag(ctx, imageConfig.Ref, newTag)
}

func (c *Client) SaveImage(ctx context.Context, imageConfig *image.ImageConfig, outputFile string) error {
	rc, err := c.wrapped.ImageSave(ctx, []string{imageConfig.Ref})
	if err != nil {
		return err
	}
	defer rc.Close()

	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, rc)
	return err
}

func (c *Client) LoadImage(ctx context.Context, inputFile string) error {
	file, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	res, err := c.wrapped.ImageLoad(ctx, file, true)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if _, err = io.Copy(c.imageResWriter, res.Body); err != nil {
		return err
	}
	return nil
}

// RunAndWait creates, starts a container and waits for it to finish.
// This is a blocking call that will not return until either:
// - The container finishes executing
// - An error occurs
// - The context is cancelled
// Use context with timeout or cancellation to control the maximum wait time.
func (c *Client) RunAndWait(ctx context.Context, containerConfig *container.ContainerConfig) error {
	if err := c.CreateContainer(ctx, containerConfig); err != nil {
		return fmt.Errorf("create container failed: %w", err)
	}

	if err := c.StartContainer(ctx, containerConfig); err != nil {
		return fmt.Errorf("start container failed: %w", err)
	}

	statusCh, errCh := c.ContainerWait(ctx, containerConfig)
	select {
	case err := <-errCh:
		return fmt.Errorf("container wait failed: %w", err)
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("container exited with non-zero code: %d", status.StatusCode)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsContainerRunning checks if a container is currently running
func (c *Client) IsContainerRunning(ctx context.Context, containerConfig *container.ContainerConfig) (bool, error) {
	container, err := c.wrapped.ContainerInspect(ctx, containerConfig.Id)
	if err != nil {
		return false, fmt.Errorf("inspect container failed: %w", err)
	}
	return container.State.Running, nil
}

// GetContainerExitCode returns the exit code of a container
func (c *Client) GetContainerExitCode(ctx context.Context, containerConfig *container.ContainerConfig) (int, error) {
	container, err := c.wrapped.ContainerInspect(ctx, containerConfig.Id)
	if err != nil {
		return 0, fmt.Errorf("inspect container failed: %w", err)
	}
	return container.State.ExitCode, nil
}

// PullImageIfNotPresent pulls an image only if it doesn't exist locally
func (c *Client) PullImageIfNotPresent(ctx context.Context, imageConfig *image.ImageConfig) error {
	_, _, err := c.wrapped.ImageInspectWithRaw(ctx, imageConfig.Ref)
	if err == nil {
		// Image exists locally
		return nil
	}

	return c.PullImage(ctx, imageConfig)
}

// GetImageSize returns the size of an image in bytes
func (c *Client) GetImageSize(ctx context.Context, imageConfig *image.ImageConfig) (int64, error) {
	img, _, err := c.wrapped.ImageInspectWithRaw(ctx, imageConfig.Ref)
	if err != nil {
		return 0, fmt.Errorf("inspect image failed: %w", err)
	}
	return img.Size, nil
}

// GetImageCreatedTime returns when the image was created
func (c *Client) GetImageCreatedTime(ctx context.Context, imageConfig *image.ImageConfig) (string, error) {
	img, _, err := c.wrapped.ImageInspectWithRaw(ctx, imageConfig.Ref)
	if err != nil {
		return "", fmt.Errorf("inspect image failed: %w", err)
	}
	return img.Created, nil
}

// IsNetworkExists checks if a network exists
func (c *Client) IsNetworkExists(ctx context.Context, networkConfig *network.NetworkConfig) (bool, error) {
	_, err := c.wrapped.NetworkInspect(ctx, networkConfig.Id, dockerNetwork.InspectOptions{})
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("network inspect failed: %w", err)
	}
	return true, nil
}

// GetNetworkContainers returns a list of container IDs connected to a network
func (c *Client) GetNetworkContainers(ctx context.Context, networkConfig *network.NetworkConfig) ([]string, error) {
	network, err := c.wrapped.NetworkInspect(ctx, networkConfig.Id, dockerNetwork.InspectOptions{})
	if err != nil {
		return nil, fmt.Errorf("network inspect failed: %w", err)
	}

	containers := make([]string, 0, len(network.Containers))
	for id := range network.Containers {
		containers = append(containers, id)
	}
	return containers, nil
}

// IsVolumeExists checks if a volume exists
func (c *Client) IsVolumeExists(ctx context.Context, volumeConfig *volume.VolumeConfig) (bool, error) {
	_, err := c.wrapped.VolumeInspect(ctx, volumeConfig.Options.Name)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("volume inspect failed: %w", err)
	}
	return true, nil
}

// GetVolumeUsage returns the size of a volume in bytes if available
func (c *Client) GetVolumeUsage(ctx context.Context, volumeConfig *volume.VolumeConfig) (int64, error) {
	vol, err := c.wrapped.VolumeInspect(ctx, volumeConfig.Options.Name)
	if err != nil {
		return 0, fmt.Errorf("volume inspect failed: %w", err)
	}
	if vol.UsageData != nil {
		return vol.UsageData.Size, nil
	}
	return 0, nil
}

// RunAsync creates and starts a container without waiting for it to finish.
// Returns a channel that will receive the container's exit error (if any).
// The channel will be closed when the container finishes.
func (c *Client) RunAsync(ctx context.Context, containerConfig *container.ContainerConfig) (<-chan error, error) {
	if err := c.CreateContainer(ctx, containerConfig); err != nil {
		return nil, fmt.Errorf("create container failed: %w", err)
	}

	if err := c.StartContainer(ctx, containerConfig); err != nil {
		return nil, fmt.Errorf("start container failed: %w", err)
	}

	resultCh := make(chan error, 1)
	statusCh, errCh := c.ContainerWait(ctx, containerConfig)

	go func() {
		defer close(resultCh)
		select {
		case err := <-errCh:
			resultCh <- fmt.Errorf("container wait failed: %w", err)
		case <-statusCh:
			resultCh <- nil
		case <-ctx.Done():
			resultCh <- ctx.Err()
		}
	}()

	return resultCh, nil
}
