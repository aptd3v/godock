package networkoptions

import (
	"fmt"

	"github.com/aptd3v/godock/pkg/godock/networkoptions/endpointoptions"
	"github.com/docker/docker/api/types/network"
)

// SetNetworkOptions is a function type for configuring options when creating a Docker network.
type SetNetworkOptions func(options *network.CreateOptions)

// Driver sets the network driver to be used when creating the Docker network.
// Use this function to specify the network driver that will manage the network's communication.
func Driver(name string) SetNetworkOptions {
	return func(options *network.CreateOptions) {
		options.Driver = name
	}
}

// Scope sets the scope of the Docker network.
// Use this function to define the network's scope, such as "local" or "global".
func Scope(scope string) SetNetworkOptions {
	return func(options *network.CreateOptions) {
		options.Scope = scope
	}
}

// EnableIPV6 sets whether IPv6 support should be enabled for the Docker network.
// Use this function to indicate if IPv6 support should be enabled for network communication.
func EnableIPV6(enable *bool) SetNetworkOptions {
	return func(options *network.CreateOptions) {
		options.EnableIPv6 = enable
	}
}

// Internal sets whether the Docker network is intended to be internal.
// Use this function to define whether the network should only be accessible within the host environment.
// If set to true, the network is restricted to communication within containers on the same host.
// If set to false (the default), the network can allow communication between containers across different hosts.
func Internal() SetNetworkOptions {
	return func(options *network.CreateOptions) {
		options.Internal = true
	}
}

// Attachable sets whether the Docker network is attachable.
// Use this function to specify if other containers can attach to this network.
func Attachable() SetNetworkOptions {
	return func(options *network.CreateOptions) {
		options.Attachable = true
	}
}

// Ingress sets whether the Docker network is an ingress network.
// Use this function to indicate if the network is the ingress network used for routing externally.
func Ingress() SetNetworkOptions {
	return func(options *network.CreateOptions) {
		options.Ingress = true
	}
}

// ConfigOnly sets whether the Docker network is a config-only network.
// Use this function to indicate if the network is only used for storing service configuration details.
func ConfigOnly() SetNetworkOptions {
	return func(options *network.CreateOptions) {
		options.ConfigOnly = true
	}
}

// ConfigFrom specifies the source which provides a network's configuration
func ConfigFrom(net fmt.Stringer) SetNetworkOptions {
	return func(options *network.CreateOptions) {
		options.ConfigFrom = &network.ConfigReference{
			Network: net.String(),
		}
	}
}

// Options sets custom options for the Docker network during creation.
// Use this function to provide additional key-value pairs for network configuration.
// These options allow you to customize specific behaviors and settings of the network.
func Options(key, value string) SetNetworkOptions {
	return func(options *network.CreateOptions) {
		if options.Options == nil {
			options.Options = map[string]string{}
		}
		options.Options[key] = value
	}
}

// Label sets labels for the Docker network during creation.
// Use this function to assign custom labels to the network for better organization and identification.
// Labels are key-value pairs that can provide metadata and context to the network.
func Label(key, value string) SetNetworkOptions {
	return func(options *network.CreateOptions) {
		if options.Labels == nil {
			options.Labels = map[string]string{}
		}
		options.Labels[key] = value
	}
}

// FOR ENDPOINTS ON CONTAINER CREATION
type SetContainerNetworkOptFn func(options *network.NetworkingConfig)

/*
Adds a networking endpoint option for the networking configuration.
*/
func Endpoint(name string, endpoint *endpointoptions.Endpoint) SetContainerNetworkOptFn {
	return func(net *network.NetworkingConfig) {
		if net.EndpointsConfig == nil {
			net.EndpointsConfig = make(map[string]*network.EndpointSettings)
		}
		net.EndpointsConfig[name] = endpoint.Settings
	}
}
