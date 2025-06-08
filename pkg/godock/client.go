package godock

import (
	"context"
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
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

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
	return c.wrapped.ContainerStart(ctx, containerConfig.Name, containerType.StartOptions{})
}

// GetContainerStats gets stats and is synchronus
// This is a blocking call and will return when the container is stopped or the context is cancelled
func (c *Client) GetContainerStats(ctx context.Context, containerConfig *container.ContainerConfig) error {

	res, err := c.wrapped.ContainerStats(ctx, containerConfig.Name, true)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if _, err := io.Copy(c.statsResWriter, res.Body); err != nil {
		return err
	}
	return nil
}

// GetContainerLogs gets logs and is synchronus
func (c *Client) GetContainerLogs(ctx context.Context, containerConfig *container.ContainerConfig) error {
	rc, err := c.wrapped.ContainerLogs(ctx, containerConfig.Id, containerType.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return err
	}
	defer rc.Close()

	if _, err := stdcopy.StdCopy(c.imageResWriter, c.imageResWriter, rc); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
func (c *Client) RemoveContainer(ctx context.Context, containerConfig *container.ContainerConfig, force bool) error {
	return c.wrapped.ContainerRemove(ctx, containerConfig.Id, containerType.RemoveOptions{
		RemoveVolumes: force,
		Force:         force,
	})
}
func (c *Client) UnpauseContainer(ctx context.Context, containerConfig *container.ContainerConfig) error {
	return c.wrapped.ContainerUnpause(ctx, containerConfig.Name)
}
func (c *Client) PauseContainer(ctx context.Context, containerConfig *container.ContainerConfig) error {
	return c.wrapped.ContainerPause(ctx, containerConfig.Name)
}
func (c *Client) RestartContainer(ctx context.Context, containerConfig *container.ContainerConfig) error {
	return c.wrapped.ContainerRestart(ctx, containerConfig.Name, containerType.StopOptions{})
}

func (c *Client) StopContainer(ctx context.Context, containerConfig *container.ContainerConfig) error {
	return c.wrapped.ContainerStop(ctx, containerConfig.Name, containerType.StopOptions{})
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
	if _, err = io.Copy(c.imageResWriter, rc); err != nil {
		return err
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
	return c.wrapped.NetworkConnect(ctx, networkConfig.Id, containerConfig.Id, nil)
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
