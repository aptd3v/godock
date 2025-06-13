package endpointoptions

import (
	"github.com/docker/docker/api/types/network"
)

type SetEndpointSettingsFn func(settings *network.EndpointSettings)

// Endpoint represents a network endpoint configuration
type Endpoint struct {
	Settings *network.EndpointSettings
}

// SetEndpointSetting applies the provided endpoint settings
func (ew *Endpoint) SetEndpointSetting(setEpSFns ...SetEndpointSettingsFn) {
	for _, set := range setEpSFns {
		if set != nil {
			set(ew.Settings)
		}
	}
}

// NewConfig creates a new endpoint configuration
func NewConfig() *Endpoint {
	return &Endpoint{
		Settings: &network.EndpointSettings{},
	}
}

/*
DriverOpt adds a driver-specific option to the endpoint settings.

Usage example:

	endpoint := endpointoptions.NewEndpoint()
	endpoint.SetEndpointSetting(
		endpointoptions.DriverOpt("com.docker.network.driver.mtu", "1500"),
	)
*/
func DriverOpt(key, value string) SetEndpointSettingsFn {
	return func(es *network.EndpointSettings) {
		if es.DriverOpts == nil {
			es.DriverOpts = make(map[string]string)
		}
		es.DriverOpts[key] = value
	}
}

/*
IPv4Address sets the IPv4 address for the endpoint.

Usage example:

	endpoint := endpointoptions.NewEndpoint()
	endpoint.SetEndpointSetting(
		endpointoptions.IPv4Address("172.20.0.2"),
	)
*/
func IPv4Address(addr string) SetEndpointSettingsFn {
	return func(es *network.EndpointSettings) {
		es.IPAddress = addr
	}
}

/*
IPv4Gateway sets the IPv4 gateway for the endpoint.

Usage example:

	endpoint := endpointoptions.NewEndpoint()
	endpoint.SetEndpointSetting(
		endpointoptions.IPv4Gateway("172.20.0.1"),
	)
*/
func IPv4Gateway(gateway string) SetEndpointSettingsFn {
	return func(es *network.EndpointSettings) {
		es.Gateway = gateway
	}
}

/*
IPv4PrefixLen sets the IPv4 subnet prefix length for the endpoint.

Usage example:

	endpoint := endpointoptions.NewEndpoint()
	endpoint.SetEndpointSetting(
		endpointoptions.IPv4PrefixLen(24), // for a /24 subnet
	)
*/
func IPv4PrefixLen(prefixLen int) SetEndpointSettingsFn {
	return func(es *network.EndpointSettings) {
		es.IPPrefixLen = prefixLen
	}
}

/*
IPv6Address sets the IPv6 address for the endpoint.

Usage example:

	endpoint := endpointoptions.NewEndpoint()
	endpoint.SetEndpointSetting(
		endpointoptions.IPv6Address("2001:db8::2"),
	)
*/
func IPv6Address(addr string) SetEndpointSettingsFn {
	return func(es *network.EndpointSettings) {
		es.GlobalIPv6Address = addr
	}
}

/*
IPv6Gateway sets the IPv6 gateway for the endpoint.

Usage example:

	endpoint := endpointoptions.NewEndpoint()
	endpoint.SetEndpointSetting(
		endpointoptions.IPv6Gateway("2001:db8::1"),
	)
*/
func IPv6Gateway(gateway string) SetEndpointSettingsFn {
	return func(es *network.EndpointSettings) {
		es.IPv6Gateway = gateway
	}
}

/*
IPv6PrefixLen sets the IPv6 subnet prefix length for the endpoint.

Usage example:

	endpoint := endpointoptions.NewEndpoint()
	endpoint.SetEndpointSetting(
		endpointoptions.IPv6PrefixLen(64), // for a /64 subnet
	)
*/
func IPv6PrefixLen(prefixLen int) SetEndpointSettingsFn {
	return func(es *network.EndpointSettings) {
		es.GlobalIPv6PrefixLen = prefixLen
	}
}

/*
MacAddress sets the MAC address for the endpoint.

Usage example:

	endpoint := endpointoptions.NewEndpoint()
	endpoint.SetEndpointSetting(
		endpointoptions.MacAddress("02:42:ac:11:00:02"),
	)
*/
func MacAddress(mac string) SetEndpointSettingsFn {
	return func(es *network.EndpointSettings) {
		es.MacAddress = mac
	}
}

/*
Links adds container links to the endpoint.
Links allow containers to discover and securely communicate with each other.

Usage example:

	endpoint := endpointoptions.NewEndpoint()
	endpoint.SetEndpointSetting(
		endpointoptions.Links("redis:cache", "postgres:db"),
	)
*/
func Links(links ...string) SetEndpointSettingsFn {
	return func(es *network.EndpointSettings) {
		if es.Links == nil {
			es.Links = make([]string, 0)
		}
		es.Links = append(es.Links, links...)
	}
}

/*
Aliases adds DNS aliases for the endpoint.
Aliases allow the container to be discovered using additional names.

Usage example:

	endpoint := endpointoptions.NewEndpoint()
	endpoint.SetEndpointSetting(
		endpointoptions.Aliases("web", "api"),
	)
*/
func Aliases(aliases ...string) SetEndpointSettingsFn {
	return func(es *network.EndpointSettings) {
		if es.Aliases == nil {
			es.Aliases = make([]string, 0)
		}
		es.Aliases = append(es.Aliases, aliases...)
	}
}

/*
NetworkID sets the ID of the network this endpoint belongs to.

Usage example:

	endpoint := endpointoptions.NewEndpoint()
	endpoint.SetEndpointSetting(
		endpointoptions.NetworkID("n1230984jf02jf"),
	)
*/
func NetworkID(id string) SetEndpointSettingsFn {
	return func(es *network.EndpointSettings) {
		es.NetworkID = id
	}
}

/*
EndpointID sets the ID for this endpoint.

Usage example:

	endpoint := endpointoptions.NewEndpoint()
	endpoint.SetEndpointSetting(
		endpointoptions.EndpointID("e2309rj203j"),
	)
*/
func EndpointID(id string) SetEndpointSettingsFn {
	return func(es *network.EndpointSettings) {
		es.EndpointID = id
	}
}

/*
IPAMConfig sets the IPAM configuration for this endpoint.
This allows you to specify static IP addresses for the endpoint.

Usage example:

	endpoint := endpointoptions.NewEndpoint()
	endpoint.SetEndpointSetting(
		endpointoptions.IPAMConfig(
			"172.20.0.2",    // IPv4 Address
			"2001:db8::2",   // IPv6 Address
			[]string{"fe80::1", "fe80::2"}, // Link-local IPs
		),
	)
*/
func IPAMConfig(ipv4 string, ipv6 string, linkLocalIPs []string) SetEndpointSettingsFn {
	return func(es *network.EndpointSettings) {
		es.IPAMConfig = &network.EndpointIPAMConfig{
			IPv4Address:  ipv4,
			IPv6Address:  ipv6,
			LinkLocalIPs: linkLocalIPs,
		}
	}
}
