package container

import (
	"github.com/aptd3v/godock/pkg/godock/containeroptions"
	"github.com/aptd3v/godock/pkg/godock/hostoptions"
	"github.com/aptd3v/godock/pkg/godock/networkoptions"
	"github.com/aptd3v/godock/pkg/godock/platformoptions"
	containerType "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// Container represents a Docker container along with its configuration options.
type ContainerConfig struct {
	Id                string
	Name              string
	Options           *containerType.Config
	HostOptions       *containerType.HostConfig
	NetworkingOptions *network.NetworkingConfig
	PlatformOptions   *v1.Platform
}

// String returns the name of the Docker container.
func (c *ContainerConfig) String() string {
	return c.Name
}

// SetHostOptions configures host-related options for the Docker container config.
// Use this method to set various host options using functions from the hostopt package.
func (c *ContainerConfig) SetHostOptions(setHOFns ...hostoptions.SetHostOptFn) {
	for _, set := range setHOFns {
		if set != nil {
			set(c.HostOptions)
		}
	}
}

// SetNetworkOptions configures network-related options for the Docker container.
// Use this method to set various network options using functions from the netopt package.
func (c *ContainerConfig) SetNetworkOptions(setNwOptFns ...networkoptions.SetContainerNetworkOptFn) {
	for _, set := range setNwOptFns {
		if set != nil {
			set(c.NetworkingOptions)
		}
	}
}

// SetOptions configures options for the Docker container.
// Use this method to set various container options using functions from the containeropt package.
func (c *ContainerConfig) SetContainerOptions(setOFns ...containeroptions.SetOptionsFns) {
	for _, set := range setOFns {
		if set != nil {
			set(c.Options)
		}
	}
}

func (c *ContainerConfig) SetPlatformOptions(setPOFns ...platformoptions.SetPlatformOptions) {
	for _, set := range setPOFns {
		if set != nil {
			set(c.PlatformOptions)
		}
	}
}

// NewConfig creates a new Container config instance with the specified name.
// The Container instance contains configuration options for creating a Docker container.
func NewConfig(name string) *ContainerConfig {
	container := &ContainerConfig{
		Id:                "",
		Name:              name,
		Options:           &containerType.Config{},
		HostOptions:       &containerType.HostConfig{},
		NetworkingOptions: &network.NetworkingConfig{},
		PlatformOptions:   &v1.Platform{},
	}

	return container
}
