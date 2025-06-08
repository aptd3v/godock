package container

import (
	"testing"

	"github.com/aptd3v/godock/pkg/godock/containeroptions"
	"github.com/aptd3v/godock/pkg/godock/hostoptions"
	"github.com/aptd3v/godock/pkg/godock/networkoptions"
	"github.com/aptd3v/godock/pkg/godock/networkoptions/endpointoptions"
	"github.com/aptd3v/godock/pkg/godock/platformoptions"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	containerName := "test-container"
	c := NewConfig(containerName)

	assert.NotNil(t, c)
	assert.Equal(t, containerName, c.Name)
	assert.Empty(t, c.Id)
	assert.NotNil(t, c.Options)
	assert.NotNil(t, c.HostOptions)
	assert.NotNil(t, c.NetworkingOptions)
	assert.NotNil(t, c.PlatformOptions)
}

func TestContainerConfig_String(t *testing.T) {
	containerName := "test-container"
	c := NewConfig(containerName)

	assert.Equal(t, containerName, c.String())
}

func TestContainerConfig_SetContainerOptions(t *testing.T) {
	c := NewConfig("test-container")

	// Test setting container options
	c.SetContainerOptions(
		containeroptions.WorkingDir("/app"),
		containeroptions.Env("KEY", "value"),
		containeroptions.Label("version", "1.0"),
		containeroptions.User("1000:1000"),
	)

	assert.Equal(t, "/app", c.Options.WorkingDir)
	assert.Contains(t, c.Options.Env, "KEY=value")
	assert.Equal(t, "1.0", c.Options.Labels["version"])
	assert.Equal(t, "1000:1000", c.Options.User)
}

func TestContainerConfig_SetContainerOptions_EdgeCases(t *testing.T) {
	c := NewConfig("test-container")

	// Test empty values
	c.SetContainerOptions(
		containeroptions.WorkingDir(""),
		containeroptions.Env("EMPTY_KEY", ""),
		containeroptions.Label("empty_label", ""),
		containeroptions.User(""),
	)

	assert.Empty(t, c.Options.WorkingDir)
	assert.Contains(t, c.Options.Env, "EMPTY_KEY=")
	assert.Equal(t, "", c.Options.Labels["empty_label"])
	assert.Empty(t, c.Options.User)

	// Test multiple environment variables
	c.SetContainerOptions(
		containeroptions.Env("KEY1", "value1"),
		containeroptions.Env("KEY2", "value2"),
	)

	assert.Contains(t, c.Options.Env, "KEY1=value1")
	assert.Contains(t, c.Options.Env, "KEY2=value2")

	// Test multiple labels
	c.SetContainerOptions(
		containeroptions.Label("label1", "value1"),
		containeroptions.Label("label2", "value2"),
	)

	assert.Equal(t, "value1", c.Options.Labels["label1"])
	assert.Equal(t, "value2", c.Options.Labels["label2"])

	// Test special characters in values
	c.SetContainerOptions(
		containeroptions.WorkingDir("/path with spaces"),
		containeroptions.Env("KEY=WITH=EQUALS", "value=with=equals"),
		containeroptions.Label("label:with:colons", "value:with:colons"),
		containeroptions.User("user:with:colons"),
	)

	assert.Equal(t, "/path with spaces", c.Options.WorkingDir)
	assert.Contains(t, c.Options.Env, "KEY=WITH=EQUALS=value=with=equals")
	assert.Equal(t, "value:with:colons", c.Options.Labels["label:with:colons"])
	assert.Equal(t, "user:with:colons", c.Options.User)
}

func TestContainerConfig_SetHostOptions(t *testing.T) {
	c := NewConfig("test-container")

	// Test setting host options
	c.SetHostOptions(
		hostoptions.Memory(1024*1024*100), // 100MB
		hostoptions.CPUQuota(50000),
		hostoptions.RestartPolicy("no", 0),
		hostoptions.PortBindings("127.0.0.1", "8080", "80"),
	)

	assert.Equal(t, int64(1024*1024*100), c.HostOptions.Memory)
	assert.Equal(t, int64(50000), c.HostOptions.CPUQuota)
	assert.Equal(t, container.RestartPolicyMode("no"), c.HostOptions.RestartPolicy.Name)
	assert.Len(t, c.HostOptions.PortBindings, 1)
}

func TestContainerConfig_SetHostOptions_EdgeCases(t *testing.T) {
	c := NewConfig("test-container")

	// Test zero values
	c.SetHostOptions(
		hostoptions.Memory(0),
		hostoptions.CPUQuota(0),
		hostoptions.RestartPolicy("", 0),
	)

	assert.Equal(t, int64(0), c.HostOptions.Memory)
	assert.Equal(t, int64(0), c.HostOptions.CPUQuota)
	assert.Equal(t, container.RestartPolicyMode("no"), c.HostOptions.RestartPolicy.Name)

	// Test negative values (should be handled by Docker daemon)
	c.SetHostOptions(
		hostoptions.Memory(-1),
		hostoptions.CPUQuota(-1),
	)

	assert.Equal(t, int64(-1), c.HostOptions.Memory)
	assert.Equal(t, int64(-1), c.HostOptions.CPUQuota)

	// Test invalid restart policy (should default to "no")
	c.SetHostOptions(
		hostoptions.RestartPolicy("invalid", 0),
	)

	assert.Equal(t, container.RestartPolicyMode("no"), c.HostOptions.RestartPolicy.Name)

	// Test multiple port bindings
	c.SetHostOptions(
		hostoptions.PortBindings("127.0.0.1", "8080", "80"),
		hostoptions.PortBindings("0.0.0.0", "9090", "90"),
	)

	assert.Len(t, c.HostOptions.PortBindings, 2)
}

func TestContainerConfig_SetNetworkOptions(t *testing.T) {
	c := NewConfig("test-container")

	// Create endpoint config
	endpoint := endpointoptions.NewConfig()
	endpoint.SetEndpointSetting(
		endpointoptions.Aliases("test-alias"),
	)

	// Test setting network options
	c.SetNetworkOptions(
		networkoptions.Endpoint("test-network", endpoint),
	)

	assert.Contains(t, c.NetworkingOptions.EndpointsConfig, "test-network")
	assert.Contains(t, c.NetworkingOptions.EndpointsConfig["test-network"].Aliases, "test-alias")
}

func TestContainerConfig_SetNetworkOptions_EdgeCases(t *testing.T) {
	c := NewConfig("test-container")

	// Test empty network name
	endpoint1 := endpointoptions.NewConfig()
	c.SetNetworkOptions(
		networkoptions.Endpoint("", endpoint1),
	)

	assert.Contains(t, c.NetworkingOptions.EndpointsConfig, "")

	// Test multiple networks
	endpoint2 := endpointoptions.NewConfig()
	endpoint2.SetEndpointSetting(
		endpointoptions.Aliases("alias1", "alias2"),
	)

	endpoint3 := endpointoptions.NewConfig()
	endpoint3.SetEndpointSetting(
		endpointoptions.Aliases("alias3"),
	)

	c.SetNetworkOptions(
		networkoptions.Endpoint("network1", endpoint2),
		networkoptions.Endpoint("network2", endpoint3),
	)

	assert.Contains(t, c.NetworkingOptions.EndpointsConfig, "network1")
	assert.Contains(t, c.NetworkingOptions.EndpointsConfig, "network2")
	assert.Contains(t, c.NetworkingOptions.EndpointsConfig["network1"].Aliases, "alias1")
	assert.Contains(t, c.NetworkingOptions.EndpointsConfig["network1"].Aliases, "alias2")
	assert.Contains(t, c.NetworkingOptions.EndpointsConfig["network2"].Aliases, "alias3")
}

func TestContainerConfig_SetPlatformOptions(t *testing.T) {
	c := NewConfig("test-container")

	// Test setting platform options
	c.SetPlatformOptions(
		platformoptions.Arch("amd64"),
		platformoptions.OS("linux"),
	)

	assert.Equal(t, "amd64", c.PlatformOptions.Architecture)
	assert.Equal(t, "linux", c.PlatformOptions.OS)
}

func TestContainerConfig_SetPlatformOptions_EdgeCases(t *testing.T) {
	c := NewConfig("test-container")

	// Test empty values
	c.SetPlatformOptions(
		platformoptions.Arch(""),
		platformoptions.OS(""),
	)

	assert.Empty(t, c.PlatformOptions.Architecture)
	assert.Empty(t, c.PlatformOptions.OS)

	// Test overwriting values
	c.SetPlatformOptions(
		platformoptions.Arch("amd64"),
		platformoptions.OS("linux"),
	)

	c.SetPlatformOptions(
		platformoptions.Arch("arm64"),
		platformoptions.OS("windows"),
	)

	assert.Equal(t, "arm64", c.PlatformOptions.Architecture)
	assert.Equal(t, "windows", c.PlatformOptions.OS)
}

func TestContainerConfig_ComplexConfiguration(t *testing.T) {
	c := NewConfig("complex-test")

	// Test a more complex configuration combining multiple options
	c.SetContainerOptions(
		containeroptions.WorkingDir("/app"),
		containeroptions.Env("NODE_ENV", "production"),
		containeroptions.Env("PORT", "3000"),
		containeroptions.Label("service", "api"),
		containeroptions.Label("version", "1.0"),
		containeroptions.Expose("3000/tcp"),
	)

	c.SetHostOptions(
		hostoptions.Memory(1024*1024*512), // 512MB
		hostoptions.CPUQuota(100000),
		hostoptions.RestartPolicy("no", 0),
		hostoptions.PortBindings("0.0.0.0", "3000", "3000"),
	)

	// Create endpoint config for network
	endpoint := endpointoptions.NewConfig()
	endpoint.SetEndpointSetting(
		endpointoptions.Aliases("api-service"),
	)

	c.SetNetworkOptions(
		networkoptions.Endpoint("backend", endpoint),
	)

	c.SetPlatformOptions(
		platformoptions.Arch("amd64"),
		platformoptions.OS("linux"),
	)

	// Assert complex configuration
	assert.Equal(t, "/app", c.Options.WorkingDir)
	assert.Contains(t, c.Options.Env, "NODE_ENV=production")
	assert.Contains(t, c.Options.Env, "PORT=3000")
	assert.Equal(t, "api", c.Options.Labels["service"])
	assert.Equal(t, "1.0", c.Options.Labels["version"])
	assert.Contains(t, c.Options.ExposedPorts, nat.Port("3000/tcp"))

	assert.Equal(t, int64(1024*1024*512), c.HostOptions.Memory)
	assert.Equal(t, int64(100000), c.HostOptions.CPUQuota)
	assert.Equal(t, container.RestartPolicyMode("no"), c.HostOptions.RestartPolicy.Name)
	assert.Len(t, c.HostOptions.PortBindings, 1)

	assert.Contains(t, c.NetworkingOptions.EndpointsConfig, "backend")
	assert.Contains(t, c.NetworkingOptions.EndpointsConfig["backend"].Aliases, "api-service")

	assert.Equal(t, "amd64", c.PlatformOptions.Architecture)
	assert.Equal(t, "linux", c.PlatformOptions.OS)
}

func TestContainerConfig_ComplexConfiguration_EdgeCases(t *testing.T) {
	c := NewConfig("complex-test-edge")

	// Test setting and overwriting options multiple times
	c.SetContainerOptions(
		containeroptions.WorkingDir("/app1"),
		containeroptions.Env("ENV1", "value1"),
		containeroptions.Label("label1", "value1"),
	)

	c.SetContainerOptions(
		containeroptions.WorkingDir("/app2"),       // Should overwrite
		containeroptions.Env("ENV2", "value2"),     // Should add
		containeroptions.Label("label1", "value2"), // Should overwrite
	)

	assert.Equal(t, "/app2", c.Options.WorkingDir)
	assert.Contains(t, c.Options.Env, "ENV1=value1")
	assert.Contains(t, c.Options.Env, "ENV2=value2")
	assert.Equal(t, "value2", c.Options.Labels["label1"])

	// Test setting host options multiple times
	c.SetHostOptions(
		hostoptions.Memory(1024*1024*100),
		hostoptions.CPUQuota(50000),
	)

	c.SetHostOptions(
		hostoptions.Memory(1024*1024*200),  // Should overwrite
		hostoptions.RestartPolicy("no", 0), // Should add
	)

	assert.Equal(t, int64(1024*1024*200), c.HostOptions.Memory)
	assert.Equal(t, int64(50000), c.HostOptions.CPUQuota)
	assert.Equal(t, container.RestartPolicyMode("no"), c.HostOptions.RestartPolicy.Name)

	// Test network options with multiple endpoints and aliases
	endpoint1 := endpointoptions.NewConfig()
	endpoint1.SetEndpointSetting(
		endpointoptions.Aliases("alias1", "alias2"),
	)

	endpoint2 := endpointoptions.NewConfig()
	endpoint2.SetEndpointSetting(
		endpointoptions.Aliases("alias3"),
	)

	c.SetNetworkOptions(
		networkoptions.Endpoint("network1", endpoint1),
	)

	c.SetNetworkOptions(
		networkoptions.Endpoint("network2", endpoint2),
	)

	assert.Contains(t, c.NetworkingOptions.EndpointsConfig, "network1")
	assert.Contains(t, c.NetworkingOptions.EndpointsConfig, "network2")
	assert.Contains(t, c.NetworkingOptions.EndpointsConfig["network1"].Aliases, "alias1")
	assert.Contains(t, c.NetworkingOptions.EndpointsConfig["network1"].Aliases, "alias2")
	assert.Contains(t, c.NetworkingOptions.EndpointsConfig["network2"].Aliases, "alias3")

	// Test platform options overwriting
	c.SetPlatformOptions(
		platformoptions.Arch("amd64"),
		platformoptions.OS("linux"),
	)

	c.SetPlatformOptions(
		platformoptions.Arch("arm64"), // Should overwrite
	)

	assert.Equal(t, "arm64", c.PlatformOptions.Architecture)
	assert.Equal(t, "linux", c.PlatformOptions.OS)
}
