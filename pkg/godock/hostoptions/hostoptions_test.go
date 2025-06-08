package hostoptions

import (
	"runtime"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-units"
	"github.com/stretchr/testify/assert"
)

func TestCapabilityManagement(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test adding capabilities
	CapAdd(AUDIT_WRITE, NET_ADMIN)(hostConfig)
	assert.Contains(t, hostConfig.CapAdd, "AUDIT_WRITE")
	assert.Contains(t, hostConfig.CapAdd, "NET_ADMIN")

	// Test dropping capabilities
	CapDrop(MAC_ADMIN, SYS_ADMIN)(hostConfig)
	assert.Contains(t, hostConfig.CapDrop, "MAC_ADMIN")
	assert.Contains(t, hostConfig.CapDrop, "SYS_ADMIN")

	// Test adding empty capabilities list
	hostConfig = &container.HostConfig{}
	CapAdd()(hostConfig)
	assert.Empty(t, hostConfig.CapAdd)

	// Test dropping empty capabilities list
	CapDrop()(hostConfig)
	assert.Empty(t, hostConfig.CapDrop)
}

func TestRestartPolicy(t *testing.T) {
	hostConfig := &container.HostConfig{}

	tests := []struct {
		name          string
		mode          string
		maxRetryCount int
		expectedMode  container.RestartPolicyMode
	}{
		{"No restart", "no", 0, "no"},
		{"Always restart", "always", 0, "always"},
		{"On-failure restart", "on-failure", 5, "on-failure"},
		{"Invalid mode", "invalid", 0, "no"},
		{"Empty mode", "", 0, "no"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RestartPolicy(tt.mode, tt.maxRetryCount)(hostConfig)
			assert.Equal(t, tt.expectedMode, hostConfig.RestartPolicy.Name)
			if tt.mode == "on-failure" {
				assert.Equal(t, tt.maxRetryCount, hostConfig.RestartPolicy.MaximumRetryCount)
			}
		})
	}
}

func TestMemorySettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test memory limit
	Memory(1024 * 1024 * 100)(hostConfig) // 100MB
	assert.Equal(t, int64(1024*1024*100), hostConfig.Memory)

	// Test memory reservation
	MemoryReservation(1024 * 1024 * 50)(hostConfig) // 50MB
	assert.Equal(t, int64(1024*1024*50), hostConfig.MemoryReservation)

	// Test memory swap
	MemorySwap(1024 * 1024 * 200)(hostConfig) // 200MB
	assert.Equal(t, int64(1024*1024*200), hostConfig.MemorySwap)

	// Test kernel memory
	KernelMemory(1024 * 1024 * 25)(hostConfig) // 25MB
	assert.Equal(t, int64(1024*1024*25), hostConfig.KernelMemory)

	// Test memory swappiness
	swappiness := int64(60)
	MemorySwappiness(&swappiness)(hostConfig)
	assert.Equal(t, &swappiness, hostConfig.MemorySwappiness)
}

func TestCPUSettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test CPU shares
	CPUShares(1024)(hostConfig)
	assert.Equal(t, int64(1024), hostConfig.CPUShares)

	// Test CPU period
	CPUPeriod(100000)(hostConfig)
	assert.Equal(t, int64(100000), hostConfig.CPUPeriod)

	// Test CPU quota
	CPUQuota(50000)(hostConfig)
	assert.Equal(t, int64(50000), hostConfig.CPUQuota)

	// Test CPUset CPUs
	CpusetCpus("0-3")(hostConfig)
	assert.Equal(t, "0-3", hostConfig.CpusetCpus)

	// Test CPUset Mems
	CpusetMems("0-1")(hostConfig)
	assert.Equal(t, "0-1", hostConfig.CpusetMems)

	// Test CPU realtime settings
	CPURealtimePeriod(100000)(hostConfig)
	assert.Equal(t, int64(100000), hostConfig.CPURealtimePeriod)

	CPURealtimeRuntime(95000)(hostConfig)
	assert.Equal(t, int64(95000), hostConfig.CPURealtimeRuntime)
}

func TestNetworkSettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test network mode
	NetworkMode("host")(hostConfig)
	assert.Equal(t, container.NetworkMode("host"), hostConfig.NetworkMode)

	NetworkMode("bridge")(hostConfig)
	assert.Equal(t, container.NetworkMode("bridge"), hostConfig.NetworkMode)

	NetworkMode("container:test")(hostConfig)
	assert.Equal(t, container.NetworkMode("container:test"), hostConfig.NetworkMode)

	NetworkMode("custom-network")(hostConfig)
	assert.Equal(t, container.NetworkMode("custom-network"), hostConfig.NetworkMode)
}

func TestPortBindings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test single port binding
	PortBindings("127.0.0.1", "8080", "80")(hostConfig)
	binding := hostConfig.PortBindings[nat.Port("80")]
	assert.Len(t, binding, 1)
	assert.Equal(t, "127.0.0.1", binding[0].HostIP)
	assert.Equal(t, "8080", binding[0].HostPort)

	// Test multiple port bindings
	PortBindings("0.0.0.0", "9090", "90")(hostConfig)
	assert.Len(t, hostConfig.PortBindings, 2)
	binding = hostConfig.PortBindings[nat.Port("90")]
	assert.Equal(t, "0.0.0.0", binding[0].HostIP)
	assert.Equal(t, "9090", binding[0].HostPort)
}

func TestMountSettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test bind mount
	Mount(MountType(mount.TypeBind), "/host/path", "/container/path", true)(hostConfig)
	assert.Len(t, hostConfig.Mounts, 1)
	assert.Equal(t, mount.TypeBind, hostConfig.Mounts[0].Type)
	assert.Equal(t, "/host/path", hostConfig.Mounts[0].Source)
	assert.Equal(t, "/container/path", hostConfig.Mounts[0].Target)
	assert.True(t, hostConfig.Mounts[0].ReadOnly)

	// Test volume mount
	Mount(MountType(mount.TypeVolume), "myvolume", "/data", false)(hostConfig)
	assert.Len(t, hostConfig.Mounts, 2)
	assert.Equal(t, mount.TypeVolume, hostConfig.Mounts[1].Type)
	assert.Equal(t, "myvolume", hostConfig.Mounts[1].Source)
	assert.Equal(t, "/data", hostConfig.Mounts[1].Target)
	assert.False(t, hostConfig.Mounts[1].ReadOnly)
}

func TestDNSSettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test DNS servers
	LookupDNS("8.8.8.8", "1.1.1.1")(hostConfig)
	assert.Contains(t, hostConfig.DNS, "8.8.8.8")
	assert.Contains(t, hostConfig.DNS, "1.1.1.1")

	// Test DNS options
	DNSOptions("use-vc", "attempts:3")(hostConfig)
	assert.Contains(t, hostConfig.DNSOptions, "use-vc")
	assert.Contains(t, hostConfig.DNSOptions, "attempts:3")

	// Test DNS search domains
	DNSSearch("example.com", "example.org")(hostConfig)
	assert.Contains(t, hostConfig.DNSSearch, "example.com")
	assert.Contains(t, hostConfig.DNSSearch, "example.org")
}

func TestSecuritySettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test privileged mode
	Privileged()(hostConfig)
	assert.True(t, hostConfig.Privileged)

	// Test security options
	SecurityOpt("label:disable")(hostConfig)
	assert.Contains(t, hostConfig.SecurityOpt, "label:disable")

	// Test no new privileges
	NoNewPrivileges()(hostConfig)
	assert.Contains(t, hostConfig.SecurityOpt, "no-new-privileges")
}

func TestResourceLimits(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test PID limits
	limit := int64(100)
	PidsLimit(&limit)(hostConfig)
	assert.Equal(t, &limit, hostConfig.PidsLimit)

	// Test Ulimits
	ulimits := []*units.Ulimit{
		{
			Name: "nofile",
			Hard: 1024,
			Soft: 1024,
		},
	}
	Ulimits(ulimits)(hostConfig)
	assert.Equal(t, ulimits, hostConfig.Ulimits)
}

func TestDeviceSettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test device mapping
	AddDevice("/dev/sda", "/dev/xvda", "rwm")(hostConfig)
	assert.Len(t, hostConfig.Devices, 1)
	assert.Equal(t, "/dev/sda", hostConfig.Devices[0].PathOnHost)
	assert.Equal(t, "/dev/xvda", hostConfig.Devices[0].PathInContainer)
	assert.Equal(t, "rwm", hostConfig.Devices[0].CgroupPermissions)

	// Test block IO settings
	BlkioWeight(500)(hostConfig)
	assert.Equal(t, uint16(500), hostConfig.BlkioWeight)

	BlkioDeviceReadBps("/dev/sda", 1024*1024)(hostConfig)
	assert.Len(t, hostConfig.BlkioDeviceReadBps, 1)
	assert.Equal(t, "/dev/sda", hostConfig.BlkioDeviceReadBps[0].Path)
	assert.Equal(t, uint64(1024*1024), hostConfig.BlkioDeviceReadBps[0].Rate)

	BlkioDeviceWriteBps("/dev/sda", 1024*1024)(hostConfig)
	assert.Len(t, hostConfig.BlkioDeviceWriteBps, 1)
	assert.Equal(t, "/dev/sda", hostConfig.BlkioDeviceWriteBps[0].Path)
	assert.Equal(t, uint64(1024*1024), hostConfig.BlkioDeviceWriteBps[0].Rate)
}

func TestMiscSettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test auto-remove
	AutoRemove()(hostConfig)
	assert.True(t, hostConfig.AutoRemove)

	// Test init
	Init()(hostConfig)
	assert.True(t, *hostConfig.Init)

	// Test storage options
	StorageOpt("size", "20G")(hostConfig)
	assert.Equal(t, "20G", hostConfig.StorageOpt["size"])

	// Test tmpfs
	Tmpfs("/tmp", "size=100m")(hostConfig)
	assert.Equal(t, "size=100m", hostConfig.Tmpfs["/tmp"])

	// Test sysctls
	sysctls := map[string]string{
		"net.ipv4.ip_forward": "1",
	}
	Sysctls(sysctls)(hostConfig)
	assert.Equal(t, sysctls, hostConfig.Sysctls)
}

func TestDeviceRequests(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test GPU device request
	DeviceRequest(
		"nvidia",
		1,
		[]string{"0"},
		[]string{"gpu", "compute"},
	)(hostConfig)

	assert.Len(t, hostConfig.DeviceRequests, 1)
	assert.Equal(t, "nvidia", hostConfig.DeviceRequests[0].Driver)
	assert.Equal(t, 1, hostConfig.DeviceRequests[0].Count)
	assert.Equal(t, []string{"0"}, hostConfig.DeviceRequests[0].DeviceIDs)
	assert.Equal(t, [][]string{{"gpu", "compute"}}, hostConfig.DeviceRequests[0].Capabilities)

	// Test multiple device requests
	DeviceRequest(
		"nvidia",
		2,
		nil,
		[]string{"utility"},
	)(hostConfig)

	assert.Len(t, hostConfig.DeviceRequests, 2)
	assert.Equal(t, 2, hostConfig.DeviceRequests[1].Count)
	assert.Nil(t, hostConfig.DeviceRequests[1].DeviceIDs)
	assert.Equal(t, [][]string{{"utility"}}, hostConfig.DeviceRequests[1].Capabilities)
}

func TestWindowsSpecificSettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test ConsoleSize
	ConsoleSize(300, 400)(hostConfig)
	if runtime.GOOS == "windows" {
		assert.Equal(t, [2]uint{300, 400}, hostConfig.ConsoleSize)
	} else {
		assert.Equal(t, [2]uint{0, 0}, hostConfig.ConsoleSize)
	}

	// Test Isolation
	Isolation("hyperv")(hostConfig)
	if runtime.GOOS == "windows" {
		assert.Equal(t, container.Isolation("hyperv"), hostConfig.Isolation)
	} else {
		assert.Equal(t, container.Isolation(""), hostConfig.Isolation)
	}
}

func TestPathSettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test ReadonlyPaths
	ReadonlyPaths("/etc", "/usr/share")(hostConfig)
	assert.Contains(t, hostConfig.ReadonlyPaths, "/etc")
	assert.Contains(t, hostConfig.ReadonlyPaths, "/usr/share")

	// Test MaskedPaths
	MaskedPaths("/proc", "/sys")(hostConfig)
	assert.Contains(t, hostConfig.MaskedPaths, "/proc")
	assert.Contains(t, hostConfig.MaskedPaths, "/sys")
}

func TestVolumeSettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test VolumeDriver
	VolumeDriver("local")(hostConfig)
	assert.Equal(t, "local", hostConfig.VolumeDriver)

	// Test VolumesFrom
	VolumesFrom("container1:rw")(hostConfig)
	assert.Contains(t, hostConfig.VolumesFrom, "container1:rw")

	// Test multiple VolumesFrom
	VolumesFrom("container2:ro")(hostConfig)
	assert.Contains(t, hostConfig.VolumesFrom, "container2:ro")

	// Test Bind
	Bind("/host:/container:ro")(hostConfig)
	assert.Contains(t, hostConfig.Binds, "/host:/container:ro")
}

func TestNamespaceSettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test UTSMode
	UTSMode("host")(hostConfig)
	assert.Equal(t, container.UTSMode("host"), hostConfig.UTSMode)

	// Test UserNSMode
	UserNSMode("host")(hostConfig)
	assert.Equal(t, container.UsernsMode("host"), hostConfig.UsernsMode)

	// Test IpcMode
	IpcMode("host")(hostConfig)
	assert.Equal(t, container.IpcMode("host"), hostConfig.IpcMode)

	// Test invalid IpcMode
	IpcMode("invalid")(hostConfig)
	assert.Equal(t, container.IpcMode("private"), hostConfig.IpcMode)

	// Test PidMode
	PidMode("host")(hostConfig)
	assert.Equal(t, container.PidMode("host"), hostConfig.PidMode)

	PidMode("container:test")(hostConfig)
	assert.Equal(t, container.PidMode("container:test"), hostConfig.PidMode)
}

func TestCgroupSettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test Cgroup
	Cgroup("cpu,memory")(hostConfig)
	assert.Equal(t, container.CgroupSpec("cpu,memory"), hostConfig.Cgroup)

	// Test CgroupParent
	CgroupParent("/my/custom/cgroup")(hostConfig)
	assert.Equal(t, "/my/custom/cgroup", hostConfig.CgroupParent)

	// Test DeviceCgroupRules
	rules := []string{"c 1:3 rwm", "a 7:* rmw"}
	DeviceCgroupRules(rules)(hostConfig)
	assert.Equal(t, rules, hostConfig.DeviceCgroupRules)
}

func TestOOMSettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test valid OomScoreAdj
	OomScoreAdj(500)(hostConfig)
	assert.Equal(t, 500, hostConfig.OomScoreAdj)

	// Test OomScoreAdj above max
	OomScoreAdj(1001)(hostConfig)
	assert.Equal(t, 0, hostConfig.OomScoreAdj)

	// Test OomScoreAdj below min
	OomScoreAdj(-1001)(hostConfig)
	assert.Equal(t, 0, hostConfig.OomScoreAdj)
}

func TestLogSettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test valid log config
	configData := map[string]string{
		"max-size": "10m",
		"max-file": "3",
	}
	LogConfig("json-file", configData)(hostConfig)
	assert.Equal(t, "json-file", hostConfig.LogConfig.Type)
	assert.Equal(t, configData, hostConfig.LogConfig.Config)

	// Test invalid log config type
	LogConfig("invalid", configData)(hostConfig)
	assert.Equal(t, "none", hostConfig.LogConfig.Type)
	assert.Nil(t, hostConfig.LogConfig.Config)
}

func TestBlkioIOpsSettings(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test BlkioDeviceReadIOps
	BlkioDeviceReadIOps("/dev/sda", 1000)(hostConfig)
	assert.Len(t, hostConfig.BlkioDeviceReadIOps, 1)
	assert.Equal(t, "/dev/sda", hostConfig.BlkioDeviceReadIOps[0].Path)
	assert.Equal(t, uint64(1000), hostConfig.BlkioDeviceReadIOps[0].Rate)

	// Test BlkioDeviceWriteIOps
	BlkioDeviceWriteIOps("/dev/sda", 1000)(hostConfig)
	assert.Len(t, hostConfig.BlkioDeviceWriteIOps, 1)
	assert.Equal(t, "/dev/sda", hostConfig.BlkioDeviceWriteIOps[0].Path)
	assert.Equal(t, uint64(1000), hostConfig.BlkioDeviceWriteIOps[0].Rate)

	// Test multiple devices
	BlkioDeviceReadIOps("/dev/sdb", 2000)(hostConfig)
	assert.Len(t, hostConfig.BlkioDeviceReadIOps, 2)
	BlkioDeviceWriteIOps("/dev/sdb", 2000)(hostConfig)
	assert.Len(t, hostConfig.BlkioDeviceWriteIOps, 2)
}

func TestContainerIDFile(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test ContainerIDFile
	ContainerIDFile("/path/to/container-id.txt")(hostConfig)
	assert.Equal(t, "/path/to/container-id.txt", hostConfig.ContainerIDFile)
}

func TestPublishAllPorts(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test PublishAllPorts
	PublishAllPorts()(hostConfig)
	assert.True(t, hostConfig.PublishAllPorts)
}

func TestReadonlyRootfs(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test ReadonlyRootfs
	ReadonlyRootfs()(hostConfig)
	assert.True(t, hostConfig.ReadonlyRootfs)
}

func TestShmSize(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test positive ShmSize
	ShmSize(67108864)(hostConfig) // 64MB
	assert.Equal(t, int64(67108864), hostConfig.ShmSize)

	// Test negative ShmSize (should not modify the value)
	prevSize := hostConfig.ShmSize
	ShmSize(-1)(hostConfig)
	assert.Equal(t, prevSize, hostConfig.ShmSize)
}

func TestRuntime(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test Runtime
	Runtime("runc")(hostConfig)
	assert.Equal(t, "runc", hostConfig.Runtime)
}

func TestRestartAlways(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test RestartAlways
	RestartAlways()(hostConfig)
	assert.Equal(t, container.RestartPolicyMode("always"), hostConfig.RestartPolicy.Name)
	assert.Equal(t, 0, hostConfig.RestartPolicy.MaximumRetryCount)
}

func TestExtraHosts(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test single extra host
	ExtraHosts("host1.local:127.0.0.1")(hostConfig)
	assert.Contains(t, hostConfig.ExtraHosts, "host1.local:127.0.0.1")

	// Test multiple extra hosts
	ExtraHosts("host2.local:127.0.0.2", "host3.local:127.0.0.3")(hostConfig)
	assert.Contains(t, hostConfig.ExtraHosts, "host2.local:127.0.0.2")
	assert.Contains(t, hostConfig.ExtraHosts, "host3.local:127.0.0.3")

	// Test empty extra hosts
	hostConfig = &container.HostConfig{}
	ExtraHosts()(hostConfig)
	assert.Empty(t, hostConfig.ExtraHosts)
}

func TestGroupAdd(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test single group
	GroupAdd("docker")(hostConfig)
	assert.Contains(t, hostConfig.GroupAdd, "docker")

	// Test multiple groups
	GroupAdd("users", "wheel")(hostConfig)
	assert.Contains(t, hostConfig.GroupAdd, "users")
	assert.Contains(t, hostConfig.GroupAdd, "wheel")

	// Test empty groups
	hostConfig = &container.HostConfig{}
	GroupAdd()(hostConfig)
	assert.Empty(t, hostConfig.GroupAdd)
}

func TestWindowsSpecificSettingsExtended(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test ConsoleSize with zero values
	ConsoleSize(0, 0)(hostConfig)
	if runtime.GOOS == "windows" {
		assert.Equal(t, [2]uint{0, 0}, hostConfig.ConsoleSize)
	} else {
		assert.Equal(t, [2]uint{0, 0}, hostConfig.ConsoleSize)
	}

	// Test Isolation with different values
	isolationTypes := []string{"process", "hyperv", "default"}
	for _, isolationType := range isolationTypes {
		Isolation(isolationType)(hostConfig)
		if runtime.GOOS == "windows" {
			assert.Equal(t, container.Isolation(isolationType), hostConfig.Isolation)
		} else {
			assert.Equal(t, container.Isolation(""), hostConfig.Isolation)
		}
	}
}

func TestNoNewPrivilegesExtended(t *testing.T) {
	hostConfig := &container.HostConfig{}

	// Test NoNewPrivileges with nil SecurityOpt
	NoNewPrivileges()(hostConfig)
	assert.Contains(t, hostConfig.SecurityOpt, "no-new-privileges")
	assert.Equal(t, 1, len(hostConfig.SecurityOpt))

	// Test NoNewPrivileges with existing SecurityOpt
	hostConfig.SecurityOpt = []string{"label:disable"}
	NoNewPrivileges()(hostConfig)
	assert.Contains(t, hostConfig.SecurityOpt, "label:disable")
	assert.Contains(t, hostConfig.SecurityOpt, "no-new-privileges")
	assert.Equal(t, 2, len(hostConfig.SecurityOpt))

	// Test NoNewPrivileges doesn't add duplicate entries
	NoNewPrivileges()(hostConfig)
	assert.Equal(t, 2, len(hostConfig.SecurityOpt), "Should not add duplicate no-new-privileges")
	assert.Contains(t, hostConfig.SecurityOpt, "no-new-privileges")
}
