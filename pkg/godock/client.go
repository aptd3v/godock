package godock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aptd3v/godock/pkg/godock/commitoptions"
	"github.com/aptd3v/godock/pkg/godock/container"
	"github.com/aptd3v/godock/pkg/godock/errdefs"
	"github.com/aptd3v/godock/pkg/godock/exec"
	"github.com/aptd3v/godock/pkg/godock/image"
	"github.com/aptd3v/godock/pkg/godock/network"
	"github.com/aptd3v/godock/pkg/godock/networkoptions/endpointoptions"
	"github.com/aptd3v/godock/pkg/godock/terminal"
	"github.com/aptd3v/godock/pkg/godock/volume"
	"github.com/docker/docker/api/types"
	containerType "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	imageType "github.com/docker/docker/api/types/image"
	dockerNetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	volumeType "github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
)

type Client struct {
	wrapped *client.Client
}

func (c *Client) ContainerCreate(ctx context.Context, containerConfig *container.ContainerConfig) error {
	if containerConfig == nil {
		return &errdefs.ValidationError{
			Field:   "containerConfig",
			Message: "container config cannot be nil",
		}
	}

	res, err := c.wrapped.ContainerCreate(
		ctx,
		containerConfig.Options,
		containerConfig.HostOptions,
		containerConfig.NetworkingOptions,
		containerConfig.PlatformOptions,
		containerConfig.Name,
	)
	if err != nil {
		if client.IsErrNotFound(err) {
			return &errdefs.ResourceNotFoundError{
				ResourceType: "image",
				ID:           containerConfig.Options.Image,
			}
		}
		// Check for conflicts
		if strings.Contains(err.Error(), "Conflict") {
			if strings.Contains(err.Error(), "is already in use") {
				return &errdefs.ResourceExistsError{
					ResourceType: "container",
					ID:           containerConfig.Name,
				}
			}
		}
		return &errdefs.ContainerError{
			ID:      containerConfig.Name,
			Op:      "create",
			Message: err.Error(),
		}
	}

	containerConfig.Id = res.ID
	return nil
}

func (c *Client) ContainerStart(ctx context.Context, containerConfig *container.ContainerConfig) error {
	if containerConfig == nil || containerConfig.Id == "" {
		return &errdefs.ValidationError{
			Field:   "containerConfig",
			Message: "container config or ID cannot be empty",
		}
	}

	err := c.wrapped.ContainerStart(ctx, containerConfig.Id, containerType.StartOptions{})
	if err != nil {
		if client.IsErrNotFound(err) {
			return &errdefs.ResourceNotFoundError{
				ResourceType: "container",
				ID:           containerConfig.Name,
			}
		}
		if strings.Contains(err.Error(), "port is already allocated") {
			return &errdefs.ResourceExistsError{
				ResourceType: "port",
				ID:           containerConfig.Name,
			}
		}
		return &errdefs.ContainerError{
			ID:      containerConfig.Name,
			Op:      "start",
			Message: err.Error(),
		}
	}
	return nil
}

// ContainerStats gets stats and is synchronus
// This is a blocking call and will return when the container is stopped or the context is cancelled
func (c *Client) ContainerStats(ctx context.Context, containerConfig *container.ContainerConfig) (io.ReadCloser, error) {
	res, err := c.wrapped.ContainerStats(ctx, containerConfig.Id, true)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

// ContainerLogs returns a ReadCloser for container logs. Caller is responsible for closing the returned reader.
func (c *Client) ContainerLogs(ctx context.Context, containerConfig *container.ContainerConfig) (io.ReadCloser, error) {
	rc, err := c.wrapped.ContainerLogs(ctx, containerConfig.Id, containerType.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return nil, err
	}

	return rc, nil
}

func (c *Client) ContainerRemove(ctx context.Context, containerConfig *container.ContainerConfig, force bool) error {
	return c.wrapped.ContainerRemove(ctx, containerConfig.Id, containerType.RemoveOptions{
		RemoveVolumes: force,
		Force:         force,
	})
}

func (c *Client) ContainerUnpause(ctx context.Context, containerConfig *container.ContainerConfig) error {
	return c.wrapped.ContainerUnpause(ctx, containerConfig.Id)
}

func (c *Client) ContainerPause(ctx context.Context, containerConfig *container.ContainerConfig) error {
	return c.wrapped.ContainerPause(ctx, containerConfig.Id)
}

func (c *Client) ContainerRestart(ctx context.Context, containerConfig *container.ContainerConfig) error {
	return c.wrapped.ContainerRestart(ctx, containerConfig.Id, containerType.StopOptions{})
}

func (c *Client) ContainerStop(ctx context.Context, containerConfig *container.ContainerConfig) error {
	return c.wrapped.ContainerStop(ctx, containerConfig.Id, containerType.StopOptions{})
}

// ContainerWait waits for a container to finish and returns a channel for status and errors
func (c *Client) ContainerWait(ctx context.Context, containerConfig *container.ContainerConfig) (<-chan containerType.WaitResponse, <-chan error) {
	return c.wrapped.ContainerWait(ctx, containerConfig.Id, containerType.WaitConditionNotRunning)
}

func (c *Client) NetworkCreate(ctx context.Context, networkConfig *network.NetworkConfig) error {
	if networkConfig == nil || networkConfig.Name == "" {
		return &errdefs.ValidationError{
			Field:   "networkConfig",
			Message: "network config or name cannot be empty",
		}
	}

	res, err := c.wrapped.NetworkCreate(ctx, networkConfig.Name, *networkConfig.Options)
	if err != nil {
		if client.IsErrNotFound(err) {
			return &errdefs.ResourceNotFoundError{
				ResourceType: "network driver",
				ID:           networkConfig.Options.Driver,
			}
		}
		return &errdefs.NetworkError{
			ID:      networkConfig.Name,
			Op:      "create",
			Message: err.Error(),
		}
	}
	networkConfig.Id = res.ID
	return nil
}

func (c *Client) VolumeCreate(ctx context.Context, volumeConfig *volume.VolumeConfig) error {
	if volumeConfig == nil || volumeConfig.Options == nil {
		return &errdefs.ValidationError{
			Field:   "volumeConfig",
			Message: "volume config cannot be nil",
		}
	}

	fmt.Printf("Creating volume with options: %+v\n", volumeConfig.Options)
	fmt.Printf("Volume labels: %+v\n", volumeConfig.Options.Labels)
	vol, err := c.wrapped.VolumeCreate(ctx, *volumeConfig.Options)
	if err != nil {
		if client.IsErrNotFound(err) {
			return &errdefs.ResourceNotFoundError{
				ResourceType: "volume driver",
				ID:           volumeConfig.Options.Driver,
			}
		}
		return &errdefs.VolumeError{
			Name:    volumeConfig.Options.Name,
			Op:      "create",
			Message: err.Error(),
		}
	}
	fmt.Printf("Created volume: %+v\n", vol)
	return nil
}

// PullImage requests the docker host to pull an image from a remote registry.
// It executes the privileged function if the operation is unauthorized and it tries one more time.
// It's up to the caller to handle the io.ReadCloser and close it properly.
func (c *Client) ImagePull(ctx context.Context, imageConfig *image.ImageConfig) (io.ReadCloser, error) {
	if imageConfig == nil || imageConfig.Ref == "" {
		return nil, &errdefs.ValidationError{
			Field:   "imageConfig",
			Message: "image config or reference cannot be empty",
		}
	}

	rc, err := c.wrapped.ImagePull(ctx, imageConfig.Ref, *imageConfig.PullOptions)
	if err != nil {
		if client.IsErrNotFound(err) {
			return nil, &errdefs.ResourceNotFoundError{
				ResourceType: "image",
				ID:           imageConfig.Ref,
			}
		}
		return nil, &errdefs.ImageError{
			Ref:     imageConfig.Ref,
			Op:      "pull",
			Message: err.Error(),
		}
	}
	return rc, nil
}

// BuildImage builds an image from a directory or a context
// If the context is not included in the image config, it will return an error
// Caller is responsible for closing the response body
func (c *Client) ImageBuild(ctx context.Context, imageConfig *image.ImageConfig) (io.ReadCloser, error) {
	res, err := c.wrapped.ImageBuild(ctx, imageConfig.BuildOptions.Context, *imageConfig.BuildOptions)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
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
		return nil, &errdefs.ConfigError{
			Field:   "client",
			Message: err.Error(),
		}
	}
	ok, err := isDaemonRunning(ctx, c)
	if err != nil {
		return nil, &errdefs.DaemonNotRunningError{
			Message: err.Error(),
		}
	}
	if !ok {
		return nil, errdefs.ErrDaemonNotRunning
	}
	return &Client{
		wrapped: c,
	}, nil
}

// Unwraps the abstracted client for use with other docker packages
func (c *Client) Unwrap() client.APIClient {
	return c.wrapped
}

// checks if the docker daemon is running by pinging it
var isDaemonRunning = func(ctx context.Context, client client.APIClient) (bool, error) {
	_, err := client.Ping(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Network Operations

func (c *Client) NetworkRemove(ctx context.Context, networkID string) error {
	return c.wrapped.NetworkRemove(ctx, networkID)
}

func (c *Client) NetworkConnect(ctx context.Context, networkConfig *network.NetworkConfig, containerConfig *container.ContainerConfig) error {
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

func (c *Client) NetworkDisconnect(ctx context.Context, networkConfig *network.NetworkConfig, containerConfig *container.ContainerConfig, force bool) error {
	return c.wrapped.NetworkDisconnect(ctx, networkConfig.Id, containerConfig.Id, force)
}

// Volume Operations

func (c *Client) VolumeRemove(ctx context.Context, name string, force bool) error {
	return c.wrapped.VolumeRemove(ctx, name, force)
}

type PruneVolumeOptionFn func(*filters.Args)

// FilterIncludeLabel adds a filter to keep volumes that have the specified label key (any value).
// Example: FilterIncludeLabel("env") keeps volumes with label "env"
func FilterIncludeLabel(key string) PruneVolumeOptionFn {
	return func(args *filters.Args) {
		args.Add("all", "true") // Enable pruning
		args.Add("label!", key) // Keep volumes with this label
	}
}

// FilterIncludeLabelValue adds a filter to keep volumes with the specified label key=value.
// Example: FilterIncludeLabelValue("env", "prod") keeps volumes with label env=prod
func FilterIncludeLabelValue(key, value string) PruneVolumeOptionFn {
	return func(args *filters.Args) {
		args.Add("all", "true")                              // Enable pruning
		args.Add("label!", fmt.Sprintf("%s=%s", key, value)) // Keep volumes with this label=value
	}
}

// FilterExcludeLabel adds a filter to keep volumes that don't have the specified label key.
// Example: FilterExcludeLabel("env") keeps volumes without label "env"
func FilterExcludeLabel(key string) PruneVolumeOptionFn {
	return func(args *filters.Args) {
		args.Add("all", "true") // Enable pruning
		args.Add("label", key)  // Keep volumes without this label
	}
}

// FilterExcludeLabelValue adds a filter to keep volumes without the specified label key=value.
// Example: FilterExcludeLabelValue("env", "prod") keeps volumes without label env=prod
func FilterExcludeLabelValue(key, value string) PruneVolumeOptionFn {
	return func(args *filters.Args) {
		args.Add("all", "true")                             // Enable pruning
		args.Add("label", fmt.Sprintf("%s=%s", key, value)) // Keep volumes without this label=value
	}
}

func (c *Client) VolumePrune(ctx context.Context, pruneVolumeOptionFns ...PruneVolumeOptionFn) (*volumeType.PruneReport, error) {
	args := filters.NewArgs()
	// Add a default filter to enable pruning of unused volumes if no other filters are provided
	if len(pruneVolumeOptionFns) == 0 {
		args.Add("all", "true")
	}
	for _, fn := range pruneVolumeOptionFns {
		if fn != nil {
			fn(&args)
		}
	}
	// Log the filter arguments
	fmt.Printf("Volume prune filter args: %+v\n", args)
	report, err := c.wrapped.VolumesPrune(ctx, args)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (c *Client) ImagePush(ctx context.Context, imageConfig *image.ImageConfig) (io.ReadCloser, error) {
	rc, err := c.wrapped.ImagePush(ctx, imageConfig.Ref, *imageConfig.PushOptions)
	if err != nil {
		return nil, err
	}
	return rc, nil
}

func (c *Client) ImageRemove(ctx context.Context, imageID string, force bool, pruneChildren bool) ([]imageType.DeleteResponse, error) {
	return c.wrapped.ImageRemove(ctx, imageID, imageType.RemoveOptions{
		Force:         force,
		PruneChildren: pruneChildren,
	})
}

func (c *Client) ImageTag(ctx context.Context, imageConfig *image.ImageConfig, newTag string) error {
	return c.wrapped.ImageTag(ctx, imageConfig.Ref, newTag)
}

func (c *Client) ImageSave(ctx context.Context, imageConfig *image.ImageConfig, outputFile string) error {
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

func (c *Client) ImageLoad(ctx context.Context, inputFile string) (io.ReadCloser, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}

	res, err := c.wrapped.ImageLoad(ctx, file, true)
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

type VolumeListOptionFn func(*volumeType.ListOptions)

func WithVolumeFilter(key, value string) VolumeListOptionFn {
	return func(opts *volumeType.ListOptions) {
		opts.Filters.Add(key, value)
	}
}

func (c *Client) VolumeList(ctx context.Context, volumeListOptionFns ...VolumeListOptionFn) (volumeType.ListResponse, error) {
	opts := volumeType.ListOptions{
		Filters: filters.NewArgs(),
	}
	for _, fn := range volumeListOptionFns {
		if fn != nil {
			fn(&opts)
		}
	}
	vols, err := c.wrapped.VolumeList(ctx, opts)
	if err != nil {
		return volumeType.ListResponse{}, fmt.Errorf("inspect volume failed: %w", err)
	}

	return vols, nil
}

type ImageListOptionFn func(*imageType.ListOptions)

// WithImageFilter adds a filter to the image list operation.
func WithImageFilter(key, value string) ImageListOptionFn {
	return func(opts *imageType.ListOptions) {
		if opts.Filters.Get(key) == nil {
			opts.Filters = filters.NewArgs()
		}
		opts.Filters.Add(key, value)
	}
}

// WithImageAll sets the all flag to true in the image list operation.
func WithImageAll(all bool) ImageListOptionFn {
	return func(opts *imageType.ListOptions) {
		opts.All = all
	}
}

// WithImageSharedSize sets the shared size flag to true in the image list operation.
func WithImageSharedSize(sharedSize bool) ImageListOptionFn {
	return func(opts *imageType.ListOptions) {
		opts.SharedSize = sharedSize
	}
}

// WithImageContainerCount sets the container count flag to true in the image list operation.
func WithImageContainerCount(containerCount bool) ImageListOptionFn {
	return func(opts *imageType.ListOptions) {
		opts.ContainerCount = containerCount
	}
}

// WithImageManifests sets the manifests flag to true in the image list operation.
func WithImageManifests(manifests bool) ImageListOptionFn {
	return func(opts *imageType.ListOptions) {
		opts.Manifests = manifests
	}
}

func (c *Client) ImageList(ctx context.Context, imageListOptionFns ...ImageListOptionFn) ([]imageType.Summary, error) {
	opts := imageType.ListOptions{
		Filters: filters.NewArgs(),
	}
	for _, fn := range imageListOptionFns {
		if fn != nil {
			fn(&opts)
		}
	}
	imgs, err := c.wrapped.ImageList(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("inspect image failed: %w", err)
	}

	return imgs, nil
}

// RunAndWait creates, starts a container and waits for it to finish.
// This is a blocking call that will not return until either:
// - The container finishes executing
// - An error occurs
// - The context is cancelled
// Use context with timeout or cancellation to control the maximum wait time.
func (c *Client) RunAndWait(ctx context.Context, containerConfig *container.ContainerConfig) error {
	if err := c.ContainerCreate(ctx, containerConfig); err != nil {
		return err
	}

	if err := c.ContainerStart(ctx, containerConfig); err != nil {
		return err
	}

	statusCh, errCh := c.ContainerWait(ctx, containerConfig)
	select {
	case err := <-errCh:
		return &errdefs.ContainerError{
			ID:      containerConfig.Name,
			Op:      "wait",
			Message: err.Error(),
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return &errdefs.ContainerError{
				ID:      containerConfig.Name,
				Op:      "run",
				Message: fmt.Sprintf("exited with code %d", status.StatusCode),
			}
		}
		return nil
	case <-ctx.Done():
		switch ctx.Err() {
		case context.DeadlineExceeded:
			return errdefs.ErrTimeout
		case context.Canceled:
			return errdefs.ErrCanceled
		default:
			return ctx.Err()
		}
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
func (c *Client) VolumeUsage(ctx context.Context, name string) (*volumeType.UsageData, error) {
	vol, err := c.wrapped.VolumeInspect(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("volume inspect failed: %w", err)
	}
	if vol.UsageData != nil {
		return vol.UsageData, nil
	}
	return nil, fmt.Errorf("volume usage data not available")
}

// RunAsync creates and starts a container without waiting for it to finish.
// Returns a channel that will receive the container's exit error (if any).
// The channel will be closed when the container finishes.
func (c *Client) RunAsync(ctx context.Context, containerConfig *container.ContainerConfig) (<-chan error, error) {
	if err := c.ContainerCreate(ctx, containerConfig); err != nil {
		return nil, fmt.Errorf("create container failed: %w", err)
	}

	if err := c.ContainerStart(ctx, containerConfig); err != nil {
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

// ContainerExecAttachTerminal attaches to a container exec command and returns a terminal session
// that can be used to interact with the command. The session handles terminal setup,
// raw mode, and cleanup automatically.
func (c *Client) ContainerExecAttachTerminal(ctx context.Context, containerConfig *container.ContainerConfig, execConfig *exec.ExecConfig) (*terminal.Session, error) {
	res, err := c.wrapped.ContainerExecCreate(ctx, containerConfig.Id, *execConfig.Options)
	execConfig.ID = res.ID
	if err != nil {
		return nil, fmt.Errorf("failed to create container exec: %w", err)
	}

	hijack, err := c.wrapped.ContainerExecAttach(ctx, res.ID, containerType.ExecAttachOptions{
		ConsoleSize: execConfig.Options.ConsoleSize,
		Tty:         execConfig.Options.Tty,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to container exec: %w", err)
	}

	// Create and return a new terminal session
	session, err := terminal.NewSession(os.Stdin, hijack.Conn, hijack.Reader)
	if err != nil {
		hijack.Close()
		return nil, fmt.Errorf("failed to create terminal session: %w", err)
	}

	return session, nil
}

// ContainerExecAttach attaches to a container exec command and returns a hijacked response
// that can be used to read the output of the exec command. It is up to the caller to close the hijacked response.
func (c *Client) ContainerExecAttach(ctx context.Context, execID string, execConfig *exec.ExecConfig) (*types.HijackedResponse, error) {

	hijack, err := c.wrapped.ContainerExecAttach(ctx, execID, containerType.ExecAttachOptions{
		ConsoleSize: execConfig.Options.ConsoleSize,
		Tty:         execConfig.Options.Tty,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach to container exec: %w", err)
	}
	return &hijack, nil
}

func (c *Client) ContainerExecCreate(ctx context.Context, containerConfig *container.ContainerConfig, execConfig *exec.ExecConfig) (string, error) {
	if containerConfig == nil || execConfig == nil {
		return "", &errdefs.ValidationError{
			Field:   "config",
			Message: "container config and exec config cannot be nil",
		}
	}

	res, err := c.wrapped.ContainerExecCreate(ctx, containerConfig.Id, *execConfig.Options)
	if err != nil {
		if client.IsErrNotFound(err) {
			return "", &errdefs.ResourceNotFoundError{
				ResourceType: "container",
				ID:           containerConfig.Id,
			}
		}
		return "", &errdefs.ExecError{
			ID:      containerConfig.Id,
			Op:      "create",
			Message: err.Error(),
		}
	}
	execConfig.ID = res.ID
	return res.ID, nil
}

func (c *Client) ContainerExecStart(ctx context.Context, containerConfig *container.ContainerConfig, execConfig *exec.ExecConfig) error {
	if execConfig == nil || execConfig.ID == "" {
		return &errdefs.ValidationError{
			Field:   "execConfig",
			Message: "exec config or ID cannot be empty",
		}
	}

	err := c.wrapped.ContainerExecStart(ctx, execConfig.ID, containerType.ExecStartOptions{
		Detach:      execConfig.Options.Detach,
		ConsoleSize: execConfig.Options.ConsoleSize,
		Tty:         execConfig.Options.Tty,
	})
	if err != nil {
		if client.IsErrNotFound(err) {
			return &errdefs.ResourceNotFoundError{
				ResourceType: "exec",
				ID:           execConfig.ID,
			}
		}
		return &errdefs.ExecError{
			ID:      execConfig.ID,
			Op:      "start",
			Message: err.Error(),
		}
	}
	return nil
}

// ContainerExecInspect returns information about a container exec command.
func (c *Client) ContainerExecInspect(ctx context.Context, execConfig *exec.ExecConfig) (*containerType.ExecInspect, error) {
	inspect, err := c.wrapped.ContainerExecInspect(ctx, execConfig.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container exec: %w", err)
	}
	return &inspect, nil
}

// ContainerExecResize resizes the TTY of a container exec command.
func (c *Client) ContainerExecResize(ctx context.Context, containerConfig *container.ContainerConfig, execConfig *exec.ExecConfig, height, width uint) error {
	return c.wrapped.ContainerExecResize(ctx, execConfig.ID, containerType.ResizeOptions{
		Height: height,
		Width:  width,
	})
}

// ContainerExport retrieves the raw contents of a container and returns them as an io.ReadCloser. It's up to the caller to close the stream.
func (c *Client) ContainerExport(ctx context.Context, containerConfig *container.ContainerConfig) (io.ReadCloser, error) {
	return c.wrapped.ContainerExport(ctx, containerConfig.Id)
}

// NetworkConnect connects a container to a network.
func (c *Client) NetworkConnectContainer(ctx context.Context, networkID string, containerID string, endpoint *endpointoptions.Endpoint) error {
	return c.wrapped.NetworkConnect(ctx, networkID, containerID, endpoint.Settings)
}

// NetworkDisconnect disconnects a container from a network.
func (c *Client) NetworkDisconnectContainer(ctx context.Context, networkID string, containerID string, force bool) error {
	return c.wrapped.NetworkDisconnect(ctx, networkID, containerID, force)
}

type NetworkInspectOptionFn func(*dockerNetwork.InspectOptions)

// WithNetworkInspectScope sets the scope of the network inspect operation.
func WithNetworkInspectScope(scope string) NetworkInspectOptionFn {
	return func(opts *dockerNetwork.InspectOptions) {
		opts.Scope = scope
	}
}

// WithNetworkInspectVerbose sets the verbose flag to true in the network inspect operation.
func WithNetworkInspectVerbose() NetworkInspectOptionFn {
	return func(opts *dockerNetwork.InspectOptions) {
		opts.Verbose = true
	}
}

func (c *Client) NetworkInspect(ctx context.Context, networkID string, networkInspectOptionFns ...NetworkInspectOptionFn) (dockerNetwork.Inspect, error) {
	opt := dockerNetwork.InspectOptions{}
	for _, fn := range networkInspectOptionFns {
		if fn != nil {
			fn(&opt)
		}
	}
	return c.wrapped.NetworkInspect(ctx, networkID, opt)
}

type NetworkListOptionFn func(*dockerNetwork.ListOptions)

func WithNetworkFilter(key, value string) NetworkListOptionFn {
	return func(opts *dockerNetwork.ListOptions) {
		opts.Filters.Add(key, value)
	}
}

func (c *Client) NetworkList(ctx context.Context, networkListOptionFns ...NetworkListOptionFn) ([]dockerNetwork.Summary, error) {
	opts := dockerNetwork.ListOptions{
		Filters: filters.NewArgs(),
	}
	for _, fn := range networkListOptionFns {
		if fn != nil {
			fn(&opts)
		}
	}
	networks, err := c.wrapped.NetworkList(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}
	return networks, nil
}

type ListContainerOptionFn func(*containerType.ListOptions)

// WithContainerFilter adds a filter to the container list operation.
func WithContainerFilter(key, value string) ListContainerOptionFn {
	return func(opts *containerType.ListOptions) {
		if opts.Filters.Get(key) == nil {
			opts.Filters = filters.NewArgs()
		}
		opts.Filters.Add(key, value)
	}
}

// WithContainerAll sets the all flag to true in the container list operation.
func WithContainerAll(all bool) ListContainerOptionFn {
	return func(opts *containerType.ListOptions) {
		opts.All = all
	}
}

// WithContainerLimit sets the limit of the container list operation.
func WithContainerLimit(limit int) ListContainerOptionFn {
	return func(opts *containerType.ListOptions) {
		opts.Limit = limit
	}
}

// WithContainerSince sets the since flag to true in the container list operation.
func WithContainerSince(since string) ListContainerOptionFn {
	return func(opts *containerType.ListOptions) {
		opts.Since = since
	}
}

// WithContainerBefore sets the before flag to true in the container list operation.
func WithContainerBefore(before string) ListContainerOptionFn {
	return func(opts *containerType.ListOptions) {
		opts.Before = before
	}
}

// WithContainerSize sets the size flag to true in the container list operation.
func WithContainerSize(size bool) ListContainerOptionFn {
	return func(opts *containerType.ListOptions) {
		opts.Size = size
	}
}

// ContainerList lists all containers. provide option functions to filter the list.
func (c *Client) ContainerList(ctx context.Context, listOptionFns ...ListContainerOptionFn) ([]types.Container, error) {
	listOpts := containerType.ListOptions{
		Filters: filters.NewArgs(),
	}
	for _, fn := range listOptionFns {
		if fn != nil {
			fn(&listOpts)
		}
	}

	containers, err := c.wrapped.ContainerList(ctx, listOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	return containers, nil
}

// ContainerStatsChan returns near realtime stats for a given container.
// It is a blocking call that will not return until either:
// - The context is cancelled
// - The container is stopped
// - An error occurs
// Use context with timeout or cancellation to control the maximum wait time.
func (c *Client) ContainerStatsChan(ctx context.Context, containerConfig *container.ContainerConfig) (<-chan ContainerStats, <-chan error) {
	statsRes, err := c.wrapped.ContainerStats(ctx, containerConfig.Id, true)
	if err != nil {
		errCh := make(chan error, 1)
		errCh <- err
		close(errCh)
		return nil, errCh
	}

	statsCh := make(chan ContainerStats, 100)
	errCh := make(chan error, 1)

	go func() {
		defer close(statsCh)
		defer close(errCh)
		defer statsRes.Body.Close()

		decoder := json.NewDecoder(statsRes.Body)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var containerStats ContainerStats
				if err := decoder.Decode(&containerStats); err != nil {
					if err != io.EOF {
						errCh <- err
					}
					return
				}
				statsCh <- containerStats
			}
		}
	}()

	return statsCh, errCh
}

// ContainerStatsOneShot gets a single stat entry from a container. It differs from `ContainerStats` in that the API should not wait to prime the stats
func (c *Client) ContainerStatsOneShot(ctx context.Context, containerConfig *container.ContainerConfig) (ContainerStats, error) {
	statsRes, err := c.wrapped.ContainerStatsOneShot(ctx, containerConfig.Id)
	if err != nil {
		return ContainerStats{}, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer statsRes.Body.Close()
	decoder := json.NewDecoder(statsRes.Body)
	var containerStats ContainerStats
	if err := decoder.Decode(&containerStats); err != nil {
		return ContainerStats{}, fmt.Errorf("failed to decode container stats: %w", err)
	}
	return containerStats, nil
}

// ImageCommit applies changes to a container and creates a new tagged image.
func (c *Client) ImageCommit(ctx context.Context, containerConfig *container.ContainerConfig, imageConfig *image.ImageConfig, commitOptions ...commitoptions.CommitOptionsFn) (string, error) {
	options := containerType.CommitOptions{}
	for _, fn := range commitOptions {
		if fn != nil {
			fn(&options)
		}
	}
	res, err := c.wrapped.ContainerCommit(ctx, containerConfig.Id, options)
	if err != nil {
		return "", fmt.Errorf("failed to commit container: %w", err)
	}
	return res.ID, nil
}

// UpdateOptionFn is a function that can be used to update a container.
type UpdateOptionFn func(*containerType.UpdateConfig)

// ContainerUpdate updates a container with new configuration.
func (c *Client) ContainerUpdate(ctx context.Context, containerConfig *container.ContainerConfig, updateOptions ...UpdateOptionFn) (*containerType.ContainerUpdateOKBody, error) {
	options := containerType.UpdateConfig{}
	for _, fn := range updateOptions {
		if fn != nil {
			fn(&options)
		}
	}

	res, err := c.wrapped.ContainerUpdate(ctx, containerConfig.Id, options)
	if err != nil {
		return nil, fmt.Errorf("failed to update container: %w", err)
	}
	return &res, nil
}

// ContainerDiff returns the changes on a container's filesystem.
func (c *Client) ContainerDiff(ctx context.Context, containerConfig *container.ContainerConfig) ([]containerType.FilesystemChange, error) {
	diff, err := c.wrapped.ContainerDiff(ctx, containerConfig.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to get container diff: %w", err)
	}
	return diff, nil
}

// ContainerKill kills a container.
func (c *Client) ContainerKill(ctx context.Context, containerConfig *container.ContainerConfig, signal string) error {
	return c.wrapped.ContainerKill(ctx, containerConfig.Id, signal)
}

// ContainerRename renames a container.
func (c *Client) ContainerRename(ctx context.Context, containerConfig *container.ContainerConfig, newName string) error {
	containerConfig.Name = newName
	return c.wrapped.ContainerRename(ctx, containerConfig.Id, newName)
}

// ContainerTop returns the top process information for a container.
func (c *Client) ContainerTop(ctx context.Context, containerConfig *container.ContainerConfig, psArgs []string) (*containerType.ContainerTopOKBody, error) {
	top, err := c.wrapped.ContainerTop(ctx, containerConfig.Id, psArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to get container top: %w", err)
	}
	return &top, nil
}

// ContainerInspect returns the JSON representation of a container. It returns docker's ContainerJSON type.
func (c *Client) ContainerInspect(ctx context.Context, containerConfig *container.ContainerConfig) (types.ContainerJSON, error) {

	inspect, err := c.wrapped.ContainerInspect(ctx, containerConfig.Id)
	if err != nil {
		return types.ContainerJSON{}, fmt.Errorf("failed to get container inspect: %w", err)
	}
	return inspect, nil
}

type PruneOptionFn func(*filters.Args)

// WithPruneFilter adds a filter to the prune operation.
func WithPruneFilter(key, value string) PruneOptionFn {
	return func(filter *filters.Args) {
		filter.Add(key, value)
	}
}

// ContainerPrune prunes containers based on the provided options.
// It returns a PruneResponse containing the space reclaimed and the containers deleted.
// It uses the filters.Args type to build the filter for the prune operation.
func (c *Client) ContainerPrune(ctx context.Context, pruneOptions ...PruneOptionFn) (*containerType.PruneReport, error) {
	filter := filters.NewArgs()
	for _, fn := range pruneOptions {
		if fn != nil {
			fn(&filter)
		}
	}
	prune, err := c.wrapped.ContainersPrune(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to prune containers: %w", err)
	}
	return &prune, nil
}

func (c *Client) ImagesPrune(ctx context.Context, pruneOptions ...PruneOptionFn) (*imageType.PruneReport, error) {
	filter := filters.NewArgs()
	for _, fn := range pruneOptions {
		if fn != nil {
			fn(&filter)
		}
	}
	prune, err := c.wrapped.ImagesPrune(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to prune images: %w", err)
	}

	return &prune, nil
}

func (c *Client) ImageHistory(ctx context.Context, imageID string) ([]imageType.HistoryResponseItem, error) {
	history, err := c.wrapped.ImageHistory(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image history: %w", err)
	}
	return history, nil
}

func (c *Client) ImageInspect(ctx context.Context, imageID string) (*types.ImageInspect, error) {
	inspect, _, err := c.wrapped.ImageInspectWithRaw(ctx, imageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get image inspect: %w", err)
	}
	return &inspect, nil
}

// ImageLoad loads an image in the docker host from the client host. It's up to the caller to close the io.ReadCloser in the ImageLoadResponse returned by this function
func (c *Client) ImageLoadFromReader(ctx context.Context, input io.Reader, quiet bool) (*imageType.LoadResponse, error) {
	rc, err := c.wrapped.ImageLoad(ctx, input, quiet)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}
	return &rc, nil
}

// ImageSave retrieves one or more images from the docker host as an io.ReadCloser. It's up to the caller to store the images and close the stream.
func (c *Client) ImageSaveToReader(ctx context.Context, imageIDs []string) (io.ReadCloser, error) {
	rc, err := c.wrapped.ImageSave(ctx, imageIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to save image: %w", err)
	}
	return rc, nil
}

// ImageSearchOptionFn is a function type that configures search options for Docker images.
type ImageSearchOptionFn func(*registry.SearchOptions)

// WithSearchLimit sets the maximum number of search results to return.
// The limit must be between 1 and 100.
func WithSearchLimit(limit int) ImageSearchOptionFn {
	return func(opts *registry.SearchOptions) {
		opts.Limit = limit
	}
}

// WithSearchFilters adds filters to the search operation.
func WithSearchFilters(filters filters.Args) ImageSearchOptionFn {
	return func(opts *registry.SearchOptions) {
		opts.Filters = filters
	}
}

// ImageSearch searches for an image on Docker Hub.
// The query parameter specifies the term to search for.
// Returns a slice of SearchResult containing the search results.
func (c *Client) ImageSearch(ctx context.Context, query string, opts ...ImageSearchOptionFn) ([]registry.SearchResult, error) {
	searchOpts := registry.SearchOptions{}
	for _, fn := range opts {
		if fn != nil {
			fn(&searchOpts)
		}
	}

	results, err := c.wrapped.ImageSearch(ctx, query, searchOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to search images: %w", err)
	}
	return results, nil
}
