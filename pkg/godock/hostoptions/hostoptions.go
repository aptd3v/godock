package hostoptions

import (
	"log"
	"runtime"
	"strings"

	"github.com/docker/docker/api/types/blkiodev"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/docker/go-units"
)

type SetHostOptFn func(options *container.HostConfig)
type Capability string

// ThrottleDevice represents a structure for rate limiting device operations
type ThrottleDevice struct {
	Path string
	Rate uint64
}

const (
	// GRANTED BY DEFAULT
	// Write records to audit log
	AUDIT_WRITE Capability = "AUDIT_WRITE"

	// Make arbitrary changes to file ownership
	CHOWN Capability = "CHOWN"

	// Bypass file read/write/execute permission checks
	DAC_OVERRIDE Capability = "DAC_OVERRIDE"

	// Bypass file ownership checks
	FOWNER Capability = "FOWNER"

	// Set process UID/GID
	FSETID Capability = "FSETID"

	// Terminate processes
	KILL Capability = "KILL"

	// Create special files
	MKNOD Capability = "MKNOD"

	// Bind to low-numbered ports
	NET_BIND_SERVICE Capability = "NET_BIND_SERVICE"

	// Use raw sockets
	NET_RAW Capability = "NET_RAW"

	// Set file capabilities
	SETFCAP Capability = "SETFCAP"

	// Set group ID
	SETGID Capability = "SETGID"

	// Set process capabilities
	SETPCAP Capability = "SETPCAP"

	// Set user ID
	SETUID Capability = "SETUID"

	// Use chroot()
	SYS_CHROOT Capability = "SYS_CHROOT"

	// NOT GRANTED BY DEFAULT

	// Configure auditing and audit rules
	AUDIT_CONTROL Capability = "AUDIT_CONTROL"

	// Read auditing and audit rules
	AUDIT_READ Capability = "AUDIT_READ"

	// Employ block devices
	BLOCK_SUSPEND Capability = "BLOCK_SUSPEND"

	// Use BPF (Berkeley Packet Filter)
	BPF Capability = "BPF"

	// Use process checkpoint/restore
	CHECKPOINT_RESTORE Capability = "CHECKPOINT_RESTORE"

	// Read files and directories
	DAC_READ_SEARCH Capability = "DAC_READ_SEARCH"

	// Lock memory
	IPC_LOCK Capability = "IPC_LOCK"

	// Become IPC namespace owner
	IPC_OWNER Capability = "IPC_OWNER"

	// Establish leases on filesystem objects
	LEASE Capability = "LEASE"

	// Set immutable attributes on files
	LINUX_IMMUTABLE Capability = "LINUX_IMMUTABLE"

	// Configure MAC (Mandatory Access Control) policy
	MAC_ADMIN Capability = "MAC_ADMIN"

	// Override MAC policy
	MAC_OVERRIDE Capability = "MAC_OVERRIDE"

	// Perform network administration tasks
	NET_ADMIN Capability = "NET_ADMIN"

	// Broadcast and listen to multicast
	NET_BROADCAST Capability = "NET_BROADCAST"

	// Access perf_event Open() hypercall
	PERFMON Capability = "PERFMON"

	// Perform admin tasks, like mount filesystems
	SYS_ADMIN Capability = "SYS_ADMIN"

	// Use reboot()
	SYS_BOOT Capability = "SYS_BOOT"

	// Load and unload kernel modules
	SYS_MODULE Capability = "SYS_MODULE"

	// Modify priority for arbitrary processes
	SYS_NICE Capability = "SYS_NICE"

	// Configure process accounting
	SYS_PACCT Capability = "SYS_PACCT"

	// Trace arbitrary processes using ptrace
	SYS_PTRACE Capability = "SYS_PTRACE"

	// Perform I/O port operations
	SYS_RAWIO Capability = "SYS_RAWIO"

	// Override resource limits
	SYS_RESOURCE Capability = "SYS_RESOURCE"

	// Set system time
	SYS_TIME Capability = "SYS_TIME"

	// Configure tty devices
	SYS_TTY_CONFIG Capability = "SYS_TTY_CONFIG"

	// Configure syslog
	SYSLOG Capability = "SYSLOG"

	// Set alarm to wake system
	WAKE_ALARM Capability = "WAKE_ALARM"
)

/*
CapAdd adds specified capabilities to the host configuration for the container.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.CapAdd(Capability("NET_ADMIN")),
	)

This function allows you to add specific Linux capabilities to the container's process, enabling controlled access to privileged actions within the container.

Note: Each call to this function adds one or more capabilities to the configuration.
*/
func CapAdd(caps ...Capability) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.CapAdd == nil {
			opt.CapAdd = make(strslice.StrSlice, 0)
		}
		for _, cap := range caps {
			opt.CapAdd = append(opt.CapAdd, string(cap))
		}
	}
}

/*
CapDrop removes specified capabilities from the host configuration for the container.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.CapDrop(Capability("MAC_ADMIN")),
	)

This function allows you to remove specific Linux capabilities from the container's process, enhancing security by limiting the privileges that the container's processes possess.

Note: Each call to this function removes one or more capabilities from the configuration.
*/
func CapDrop(caps ...Capability) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.CapDrop == nil {
			opt.CapDrop = make(strslice.StrSlice, 0)
		}
		for _, cap := range caps {
			opt.CapDrop = append(opt.CapDrop, string(cap))
		}
	}
}

/*
RestartPolicy adds a restart policy to the host configuration.

Valid restart policy options:
- "no" (default policy)
- "on-failure"[:maxRetryCount]
- "always"

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.RestartPolicy("on-failure", 5),
	)

This function allows you to specify a restart policy for the container's behavior upon exit or failure.
Choose from predefined restart policies: "no", "on-failure", "always", or "unless-stopped", optionally with a maximum retry count.

Note: Calling this function sets the restart policy and, if applicable, the maximum retry count for the container.
*/
func RestartPolicy(mode string, maxRetryCount int) SetHostOptFn {
	policyMode := container.RestartPolicyDisabled
	switch mode {
	case "no":
		break
	case "on-failure":
		policyMode = container.RestartPolicyOnFailure
	case "always":
		policyMode = container.RestartPolicyAlways
	default:
		log.Printf("%s is not a valid policy defaulting to RestartPolicyDisabled aka 'no'", mode)
		policyMode = container.RestartPolicyDisabled
	}
	return func(opt *container.HostConfig) {
		opt.RestartPolicy = container.RestartPolicy{
			Name:              policyMode,
			MaximumRetryCount: maxRetryCount,
		}
	}
}

/*
Memory sets a memory limit (in bytes) for the container in the host configuration.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.Memory(int64(512 * 1024 * 1024)), // Set memory limit to 512MB
	)

This function allows you to specify a memory limit for the container's processes. The container's memory usage will be restricted
to the specified limit, preventing excessive memory consumption.

Note: Calling this function sets the memory limit for the container in bytes.
*/
func Memory(memory int64) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.Memory = memory
	}
}

/*
RestartAlways adds a restart policy that ensures the container is always restarted upon exit.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.RestartAlways(),
	)

This function sets up a restart policy in the host configuration that instructs Docker to always restart the container
automatically when it exits, ensuring continuous availability of the service.

Note: Calling this function enables the "always" restart policy for the container.
*/
func RestartAlways() SetHostOptFn {
	return RestartPolicy("always", 0)
}

/*
AutoRemove adds an auto-remove option to the host configuration, ensuring the container is removed after it exits.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.AutoRemove(),
	)

This function sets the auto-remove flag in the host configuration. When enabled, the container will be automatically removed
once it exits, helping to clean up resources and avoid manual cleanup operations.

Note: Calling this function enables the auto-remove behavior for the container.
*/
func AutoRemove() SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.AutoRemove = true
	}
}

/*
PortBindings sets up port mappings between the host and the container in the host configuration.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.PortBindings("0.0.0.0", "8080", "80"),
	)

This function allows you to specify port mappings for forwarding traffic between the host and the container.
You can map a specific host IP address and port to a container port. The host IP can be "0.0.0.0" to bind to all available interfaces.

Note: Each call to this function adds a port mapping configuration to the host configuration.
*/
func PortBindings(hostIP, hostPort, containerPort string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.PortBindings == nil {
			opt.PortBindings = make(nat.PortMap)
		}
		opt.PortBindings[nat.Port(containerPort)] = []nat.PortBinding{
			{
				HostIP:   hostIP,
				HostPort: hostPort,
			},
		}
	}
}

/*
MountType is constant for the type of mount

	    "bind"
		// TypeVolume is the type for remote storage volumes
		"volume"
		// TypeTmpfs is the type for mounting tmpfs
		"tmpfs"
		// TypeNamedPipe is the type for mounting Windows named pipes
		"npipe"
*/
type MountType mount.Type

/*
Mount configures a volume mount between the host and the container in the host configuration.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.Mount(hostoptions.MountType., "/host/source", "/container/target", true),
	)

This function allows you to specify volume mounts for sharing files or directories between the host and the container.
You can choose the mount type from predefined options using the MountType enum, such as Bind, Volume, or Tmpfs.

Note: Each call to this function adds a volume mount configuration to the host configuration.
*/
func Mount(mountType MountType, source, target string, readOnly bool) SetHostOptFn {

	return func(opt *container.HostConfig) {
		if opt.Mounts == nil {
			opt.Mounts = make([]mount.Mount, 0)
		}

		opt.Mounts = append(opt.Mounts, mount.Mount{
			Type:     mount.Type(mountType),
			Source:   source,
			Target:   target,
			ReadOnly: readOnly,
		})
	}
}

/*
LookupDNS adds a list of custom DNS servers to the host configuration for the container.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.LookupDNS("8.8.8.8", "1.1.1.1"),
	)

This function allows you to specify additional DNS server IP addresses that the container's processes can use for DNS lookups.
Custom DNS servers can be used to override the default DNS server configuration for the container.

Note: Each call to this function adds one or more custom DNS server IP addresses to the configuration.
*/
func LookupDNS(dns ...string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.DNS == nil {
			opt.DNS = make([]string, 0)
		}
		opt.DNS = append(opt.DNS, dns...)
	}
}

/*
DNSOptions adds a list of DNS resolver options to the host configuration for the container.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.DNSOptions("use-vc", "attempts:1"),
	)

This function allows you to specify additional DNS resolver options that the container's processes can use for customizing DNS resolution behavior.
DNS resolver options modify how DNS queries are performed, such as enabling validation of responses or limiting the number of attempts.

Note: Each call to this function adds one or more DNS resolver options to the configuration.
*/
func DNSOptions(dnsOption ...string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.DNSOptions == nil {
			opt.DNSOptions = make(strslice.StrSlice, 0)
		}
		opt.DNSOptions = append(opt.DNSOptions, dnsOption...)
	}
}

/*
DNSSearch adds a list of DNS search domains to the host configuration for the container.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.DNSSearch("example.com", "example.org"),
	)

This function allows you to specify additional DNS search domains that the container's processes can use for name resolution.
DNS search domains are used to complete unqualified domain names when performing DNS lookups within the container.

Note: Each call to this function adds one or more DNS search domains to the configuration.
*/
func DNSSearch(search ...string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.DNSSearch == nil {
			opt.DNSSearch = make(strslice.StrSlice, 0)
		}
		opt.DNSSearch = append(opt.DNSSearch, search...)
	}
}

/*
ExtraHosts adds a list of custom host-to-IP mappings to the host configuration for the container.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.ExtraHosts("docker.io", "docker.com"),
	)

This function allows you to specify additional host-to-IP mappings that the container's processes can use for DNS lookups.
These mappings can be used to override DNS resolutions for specific hosts within the container.

Note: Each call to this function adds one or more custom host mappings to the configuration.
*/
func ExtraHosts(extraHosts ...string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.ExtraHosts == nil {
			opt.ExtraHosts = make([]string, 0)
		}
		opt.ExtraHosts = append(opt.ExtraHosts, extraHosts...)
	}
}

/*
GroupAdd adds supplementary groups to the host configuration for the container.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.GroupAdd("wheel"),
	)

This function allows you to specify additional supplementary groups for the container's processes.
Supplementary groups provide access to shared resources or permissions that are granted based on group membership.

Note: Each call to this function adds one or more supplementary groups to the configuration.
*/
func GroupAdd(group ...string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.GroupAdd == nil {
			opt.GroupAdd = make(strslice.StrSlice, 0)
		}
		opt.GroupAdd = append(opt.GroupAdd, group...)
	}
}

/*
Bind adds volume bindings to the host configuration for the container.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.Bind("/host/path:/container/path:ro"),
	)

This function allows you to specify volume bindings between the host and the container.
Volume bindings enable sharing files and directories between the host and the container.
The format for each binding is "/host/path:/container/path:options", where "options" can include
"ro" for read-only access or "rw" for read-write access.

Note: Each call to this function adds a new volume binding to the configuration.
*/
func Bind(bind string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.Binds == nil {
			opt.Binds = make([]string, 0)
		}
		opt.Binds = append(opt.Binds, bind)
	}
}

/*
LogConfig adds a log configuration to the host configuration for the container. The default log type is "none".

Usage example:

	logConfigData := map[string]string{
		"max-size": "10m", // Maximum log file size
		"max-file": "3",   // Maximum number of log files to retain
	}
	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.LogConfig("json-file", logConfigData),
	)

This function allows you to specify the log configuration for the container's logging output.
You can choose from various log types, such as "json-file", "syslog", "journald", "gelf", "fluentd",
"awslogs", "splunk", "etwlogs", or "none".

The log configuration can include additional data specific to the chosen log type. Pass the
`configData` parameter as a map[string]string containing the configuration details for the selected log type.

Note: If an unsupported log type is provided, the function will set the log type to "none" by default.
*/
func LogConfig(configType string, configData map[string]string) SetHostOptFn {
	switch configType {
	case "json-file", "syslog", "journald", "gelf", "fluentd", "awslogs", "splunk", "etwlogs", "none":
		return func(opt *container.HostConfig) {
			opt.LogConfig = container.LogConfig{
				Type:   configType,
				Config: configData,
			}
		}
	default:
		return func(opt *container.HostConfig) {
			opt.LogConfig = container.LogConfig{
				Type: "none",
			}
		}
	}
}

/*
UTSMode sets the UTS (Unix Timesharing System) namespace mode to be used for the container in the host configuration.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.UTSMode("host"),
	)

This function allows you to specify the UTS namespace mode for the container.
The UTS namespace isolates the nodename and domain name identifiers between the host and containers.
The available modes are "host" and "container".

Note: The effect of this function depends on the container runtime environment and host configuration.
Select the appropriate UTS namespace mode based on your isolation and naming requirements.
*/
func UTSMode(mode string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.UTSMode = container.UTSMode(mode)
	}
}

/*
UserNSMode sets the user namespace mode to be used for the container in the host configuration.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.UserNSMode("host"),
	)

This function allows you to specify the user namespace mode for the container.
The user namespace provides isolation for user and group IDs between the host and containers.
The available modes are "host", "private", "shareable", and "container".

Note: The effect of this function depends on the container runtime environment and host configuration.
Make sure to use the appropriate user namespace mode that suits your security and isolation requirements.
*/
func UserNSMode(mode string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.UsernsMode = container.UsernsMode(mode)
	}
}

/*
ShmSize sets the size of the shared memory file system (/dev/shm) for the container in the host configuration.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.ShmSize(int64(67108864)),
	)

This function allows you to specify the size of the shared memory filesystem used by processes within the container.
By default, if omitted, the system allocates 64MB for /dev/shm.

Note: If a negative size is provided, the function will have no effect, and the system's default size will be used.
*/
func ShmSize(size int64) SetHostOptFn {
	if size < 0 {
		return func(opt *container.HostConfig) {
			// Do not modify, use system default
		}
	}
	return func(opt *container.HostConfig) {
		opt.ShmSize = size
	}
}

/*
Runtime sets the runtime to be used for this container in the host configuration.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.Runtime("runc"),
	)

This function allows you to specify the container runtime that will be used to execute the processes
within the container. It is applied to the host configuration.

Note: The impact of this function may vary depending on the container runtime environment.
Make sure to use a runtime that is compatible with the host and container's requirements.
*/
func Runtime(runtime string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.Runtime = runtime
	}
}

/*
ConsoleSize sets the initial console size for the container's terminal in the host configuration.
This function is intended for use in Windows environments.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.ConsoleSize(uint(300), uint(300)),
	)

Note: This function is applicable only in Windows environments. It sets the initial dimensions
of the container's terminal console as a [height, width] array. For non-Windows platforms,
it has no effect and leaves the host configuration unchanged.
*/
func ConsoleSize(height uint, width uint) SetHostOptFn {
	if runtime.GOOS != "windows" {
		return func(opt *container.HostConfig) {
			// Do nothing for non-Windows platforms
		}
	}
	return func(opt *container.HostConfig) {
		opt.ConsoleSize = [2]uint{height, width}
	}
}

/*
Isolation sets the isolation mode to be used for the container in the host configuration.
This function is intended for use in Windows environments.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.Isolation("hyperv"),
	)

Note: This function applies isolation settings only in Windows environments. For other operating systems,
it has no effect and leaves the host configuration unchanged.
*/
func Isolation(isolation string) SetHostOptFn {
	if runtime.GOOS != "windows" {
		return func(opt *container.HostConfig) {
			// Do nothing for non-Windows platforms
		}
	}
	return func(opt *container.HostConfig) {
		opt.Isolation = container.Isolation(isolation)
	}
}

/*
ReadonlyPaths adds a list of files or directories within the container's
filesystem that you want to mark as read-only.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.ReadonlyPaths("/etc", "/usr/share"),
	)

This function allows you to specify a set of paths within the container's filesystem that should be marked as read-only.
When marked as read-only, changes to these paths are not allowed, enhancing data integrity and security.

Note: Calling this function adds one or more read-only paths to the container's configuration.
*/
func ReadonlyPaths(paths ...string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.ReadonlyPaths == nil {
			opt.ReadonlyPaths = make([]string, 0)
		}
		opt.ReadonlyPaths = append(opt.ReadonlyPaths, paths...)
	}
}

/*
Adds a list of paths to be masked inside the container in the host configuration (this overrides the default set of paths).

When you use MaskedPaths, the specified files or directories will not be visible or accessible from within the container,
even if they exist in the underlying image. This can help prevent accidental or intentional modifications to critical files
or directories within the container.

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.MaskPaths("/proc", "/sys"),
	)
*/
func MaskedPaths(maskPaths ...string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.MaskedPaths == nil {
			opt.MaskedPaths = make([]string, 0)
		}
		opt.MaskedPaths = append(opt.MaskedPaths, maskPaths...)
	}
}

/*
Adds a network mode to the host configuration.
Accepted values are:
- "bridge": Use Docker's default bridge network
- "host": Use the host's network stack
- "none": No network access
- "container:<name|id>": Use another container's network namespace
- "<network-name>": Connect to a user-defined network

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		// Use host network
		hostoptions.NetworkMode("host"),
		// Or connect to a user-defined network
		hostoptions.NetworkMode("my-custom-network"),
		// Or share network namespace with another container
		hostoptions.NetworkMode("container:another-container"),
	)
*/
func NetworkMode(mode string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		// Handle container network namespace sharing
		if strings.HasPrefix(mode, "container:") {
			opt.NetworkMode = container.NetworkMode(mode)
			return
		}

		// Handle standard modes
		switch mode {
		case "bridge", "host", "none":
			opt.NetworkMode = container.NetworkMode(mode)
		default:
			// Any other value is treated as a custom network name
			opt.NetworkMode = container.NetworkMode(mode)
		}
	}
}

/*
Adds a volume driver option to the host configuration.

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.VolumeDriver("local"),
	)
*/
func VolumeDriver(driver string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.VolumeDriver = driver
	}
}

/*
VolumesFrom adds a list of volumes to inherit from another container, specified in the form <container name>[:<ro|rw>].

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.VolumesFrom("my_other_container:rw"),
	)

This function allows you to specify a list of volumes to inherit from another container. The format is <container name>[:<ro|rw>],
where <container name> is the name of the source container and ":rw" or ":ro" specifies whether the volume should be mounted read-write or read-only.

Note: Calling this function adds one or more volume sources to the container's configuration.
*/
func VolumesFrom(from string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.VolumesFrom == nil {
			opt.VolumesFrom = make([]string, 0)
		}
		opt.VolumesFrom = append(opt.VolumesFrom, from)
	}
}

/*
Adds a IPC namespace to use for the container in the host configuration
the default value is "private"
IPC sharing mode for the container. Possible values are:

	"none"
	"private"
	"shareable"
	"container"
	"host"

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.IPCMode("host"),
	)
*/
func IpcMode(mode string) SetHostOptFn {
	switch mode {
	case "none", "private", "shareable", "container", "host":
		return func(opt *container.HostConfig) {
			opt.IpcMode = container.IpcMode(mode)
		}
	default:
		return func(opt *container.HostConfig) {
			opt.IpcMode = container.IpcMode("private")
		}
	}
}

/*
Adds a Cgroup to use for the container in the host configuration

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.Cgroup("cpu,memory"),
	)
*/
func Cgroup(cgroup string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.Cgroup = container.CgroupSpec(cgroup)

	}
}

/*
Sets an integer value containing the score given to the container in order to
tune OOM killer preferences to the host configuration.
valid values are integers in the range of -1000 to 1000

The OomScoreAdj is a kernel-level mechanism in Linux that allows you to adjust
the OOM (Out-of-Memory) killer score of a process. The OOM killer is a part of
the Linux kernel that's responsible for selecting and killing processes when
the system runs out of available memory to prevent the entire system from
becoming unresponsive due to memory exhaustion.

The OOM score of a process determines its likelihood of being killed by the
OOM killer. A lower OOM score indicates a higher priority for being killed,
while a higher OOM score indicates a lower priority.

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.OomScoreAdj(100),
	)
*/
func OomScoreAdj(score int) SetHostOptFn {
	if score < -1000 || score > 1000 {
		return func(opt *container.HostConfig) {
			opt.OomScoreAdj = 0
		}
	}
	return func(opt *container.HostConfig) {
		opt.OomScoreAdj = score
	}
}

/*
Sets the PID mode to the host configuration.

"host": In this mode, the container uses the same
PID namespace as the host system. This means the
processes inside the container are not isolated
from the hosts processes. This mode can be useful
when you need processes in the container to interact
with or manage host-level processes directly.

"container:<container_id>": In this mode, the container
shares the PID namespace with another container specified
by its ID. This mode allows processes in both containers
to see each other's processes as if they were in the same
PID namespace.

(empty string, default): In this mode, the container
gets its own isolated PID namespace, which is the default behavior.

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.PidMode("host"),
	)
*/
func PidMode(mode string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.PidMode = container.PidMode(mode)
	}
}

/*
Publishes all ports in the host configuration
When this flag is enabled, all exposed ports in the
container are automatically mapped to random ports on
the host system. This allows external systems to access
services running inside the container without explicitly
specifying port mappings.

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.PublishAllPorts(),
	)
*/
func PublishAllPorts() SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.PublishAllPorts = true
	}
}

/*
Adds options to mount the container's root filesystem as read only.

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.ReadOnlyRootfs(),
	)
*/
func ReadonlyRootfs() SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.ReadonlyRootfs = true
	}
}

/*
Adds a list of string values to customize labels for MLS systems, such as SELinux.

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.SecurityOpts("label:disable"),
	)
*/
func SecurityOpt(opts ...string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.SecurityOpt == nil {
			opt.SecurityOpt = make([]string, 0)
		}
		opt.SecurityOpt = append(opt.SecurityOpt, opts...)
	}
}

/*
Adds storage driver options per container to the host configuration.

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.StorageOpt("size", "20G"),
	)
*/
func StorageOpt(key, value string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.StorageOpt == nil {
			opt.StorageOpt = make(map[string]string)
		}
		opt.StorageOpt[key] = value
	}
}

/*
Adds to a map of tmpfs (mounts) used for the container

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.Tmpfs("size", "100m"), // Set the size limit
	)
*/
func Tmpfs(key, value string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.Tmpfs == nil {
			opt.Tmpfs = make(map[string]string)
		}
		opt.Tmpfs[key] = value
	}
}

/*
Sets the Privileged mode to the host configuration which allows the following:

1. Access to All Devices: Containers in privileged mode can access all devices on the host system, which includes raw disk devices and hardware devices.

2. Capability to Modify Network Configuration: Containers can modify network settings and configurations, potentially affecting the host network and other containers.

3. Ability to Load Kernel Modules: Containers can load and unload kernel modules, which can have a broad impact on the host system's kernel.

4. Bypassing User and Namespace Isolation: Containers in privileged mode have access to the host's user namespace and can potentially perform actions that would require higher privileges.

5. Access to Hosts Process Space: Containers can interact with processes running on the host system.

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
	hostoptions.Privileged(),
	)
*/
func Privileged() SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.Privileged = true
	}
}

/*
Adds a device to the host configuration.

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.AddDevice("/dev/net/tun", "/dev/net/tun", "rwm"),
	)
*/
func AddDevice(device string, pathInContainer string, permissions string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.Devices == nil {
			opt.Devices = make([]container.DeviceMapping, 0)
		}
		opt.Devices = append(opt.Devices, container.DeviceMapping{
			PathOnHost:        device,
			PathInContainer:   pathInContainer,
			CgroupPermissions: permissions,
		})
	}
}

/*
Adds a containerIDFile to the host configuration.
After running this command, the /path/to/container-id.txt file will contain the ID of the started container.

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.ContainerIDFile("/path/to/container-id.txt"),
	)
*/
func ContainerIDFile(containerIDFile string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.ContainerIDFile = containerIDFile
	}
}

/*
CPUShares sets the CPU shares (relative weight) for the container
*/
func CPUShares(shares int64) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.CPUShares = shares
	}
}

/*
CPUPeriod sets the CPU CFS (Completely Fair Scheduler) period
*/
func CPUPeriod(period int64) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.CPUPeriod = period
	}
}

/*
CPUQuota sets the CPU CFS (Completely Fair Scheduler) quota
*/
func CPUQuota(quota int64) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.CPUQuota = quota
	}
}

/*
CpusetCpus sets the CPUs in which execution is allowed
*/
func CpusetCpus(cpus string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.CpusetCpus = cpus
	}
}

/*
MemoryReservation sets the memory soft limit
*/
func MemoryReservation(memory int64) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.MemoryReservation = memory
	}
}

/*
MemorySwap sets the total memory limit (memory + swap)
*/
func MemorySwap(memorySwap int64) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.MemorySwap = memorySwap
	}
}

/*
NoNewPrivileges disables new privileges from being acquired
*/
func NoNewPrivileges() SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.SecurityOpt == nil {
			opt.SecurityOpt = make([]string, 0)
		}
		opt.SecurityOpt = append(opt.SecurityOpt, "no-new-privileges")
	}
}

/*
Ulimits sets ulimit options
*/
func Ulimits(ulimits []*units.Ulimit) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.Ulimits = ulimits
	}
}

/*
Init runs an init inside the container
*/
func Init() SetHostOptFn {
	return func(opt *container.HostConfig) {
		t := true
		opt.Init = &t
	}
}

/*
CPURealtimePeriod sets the CPU real-time period in microseconds.
This option is only applicable when running containers on operating systems
that support CPU real-time scheduler.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.CPURealtimePeriod(100000),
	)
*/
func CPURealtimePeriod(period int64) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.CPURealtimePeriod = period
	}
}

/*
CPURealtimeRuntime sets the CPU real-time runtime in microseconds.
This option is only applicable when running containers on operating systems
that support CPU real-time scheduler.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.CPURealtimeRuntime(95000),
	)
*/
func CPURealtimeRuntime(runtime int64) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.CPURealtimeRuntime = runtime
	}
}

/*
CpusetMems sets the memory nodes in which execution is allowed.
Only effective on NUMA systems.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.CpusetMems("0-1"),
	)
*/
func CpusetMems(mems string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.CpusetMems = mems
	}
}

/*
MemorySwappiness tunes container memory swappiness (0 to 100).
- A value of 0 turns off anonymous page swapping.
- A value of 100 sets the host's swappiness value.
- Values between 0 and 100 modify the swappiness level accordingly.

Usage example:

	swappiness := int64(60)
	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.MemorySwappiness(&swappiness),
	)
*/
func MemorySwappiness(swappiness *int64) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.MemorySwappiness = swappiness
	}
}

/*
KernelMemory sets the kernel memory limit in bytes.
This is the hard limit for kernel memory that cannot be swapped out.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.KernelMemory(int64(50 * 1024 * 1024)), // 50MB kernel memory
	)
*/
func KernelMemory(memory int64) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.KernelMemory = memory
	}
}

/*
PidsLimit sets the container's PIDs limit.
Set to -1 for unlimited PIDs.

Usage example:

	limit := int64(100)
	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.PidsLimit(&limit),
	)
*/
func PidsLimit(limit *int64) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.PidsLimit = limit
	}
}

/*
BlkioWeight sets the block IO weight (relative weight) for the container.
Weight is a value between 10 and 1000.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.BlkioWeight(500),
	)
*/
func BlkioWeight(weight uint16) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.BlkioWeight = weight
	}
}

/*
BlkioDeviceReadBps sets the rate at which a device can be read in bytes per second.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.BlkioDeviceReadBps("/dev/sda", 1024*1024), // 1MB/s
	)
*/
func BlkioDeviceReadBps(devicePath string, rate uint64) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.BlkioDeviceReadBps == nil {
			opt.BlkioDeviceReadBps = make([]*blkiodev.ThrottleDevice, 0)
		}
		opt.BlkioDeviceReadBps = append(opt.BlkioDeviceReadBps, &blkiodev.ThrottleDevice{
			Path: devicePath,
			Rate: rate,
		})
	}
}

/*
BlkioDeviceWriteBps sets the rate at which a device can be written in bytes per second.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.BlkioDeviceWriteBps("/dev/sda", 1024*1024), // 1MB/s
	)
*/
func BlkioDeviceWriteBps(devicePath string, rate uint64) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.BlkioDeviceWriteBps == nil {
			opt.BlkioDeviceWriteBps = make([]*blkiodev.ThrottleDevice, 0)
		}
		opt.BlkioDeviceWriteBps = append(opt.BlkioDeviceWriteBps, &blkiodev.ThrottleDevice{
			Path: devicePath,
			Rate: rate,
		})
	}
}

/*
BlkioDeviceReadIOps sets the rate at which read operations can be performed on a device.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.BlkioDeviceReadIOps("/dev/sda", 1000), // 1000 IOPS
	)
*/
func BlkioDeviceReadIOps(devicePath string, rate uint64) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.BlkioDeviceReadIOps == nil {
			opt.BlkioDeviceReadIOps = make([]*blkiodev.ThrottleDevice, 0)
		}
		opt.BlkioDeviceReadIOps = append(opt.BlkioDeviceReadIOps, &blkiodev.ThrottleDevice{
			Path: devicePath,
			Rate: rate,
		})
	}
}

/*
BlkioDeviceWriteIOps sets the rate at which write operations can be performed on a device.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.BlkioDeviceWriteIOps("/dev/sda", 1000), // 1000 IOPS
	)
*/
func BlkioDeviceWriteIOps(devicePath string, rate uint64) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.BlkioDeviceWriteIOps == nil {
			opt.BlkioDeviceWriteIOps = make([]*blkiodev.ThrottleDevice, 0)
		}
		opt.BlkioDeviceWriteIOps = append(opt.BlkioDeviceWriteIOps, &blkiodev.ThrottleDevice{
			Path: devicePath,
			Rate: rate,
		})
	}
}

/*
Sysctls sets namespaced kernel parameters (sysctls) in the container.
These parameters can be used to tune and configure container behavior.

Usage example:

	sysctls := map[string]string{
		"net.ipv4.ip_forward": "1",
	}
	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.Sysctls(sysctls),
	)
*/
func Sysctls(sysctls map[string]string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.Sysctls = sysctls
	}
}

/*
DeviceCgroupRules sets a list of cgroup rules to allow the container to access devices.
The rules are in the format specified by the Linux kernel documentation.

Usage example:

	rules := []string{
		"c 1:3 rwm",
		"a 7:* rmw",
	}
	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.DeviceCgroupRules(rules),
	)
*/
func DeviceCgroupRules(rules []string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.DeviceCgroupRules = rules
	}
}

/*
CgroupParent sets the parent cgroup for the container.
This allows for resource sharing and limits inheritance from a parent cgroup.

Usage example:

	myContainer := container.NewConfig("my_container")
	myContainer.SetHostOptions(
		hostoptions.CgroupParent("/my/custom/cgroup"),
	)
*/
func CgroupParent(parent string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		opt.CgroupParent = parent
	}
}

/*
DeviceRequest adds a device request for the container.
This is commonly used for requesting access to GPUs or other specialized hardware.

Usage example:

	myContainer := container.NewConfig("my_container")
	// Request a specific NVIDIA GPU
	myContainer.SetHostOptions(
		hostoptions.DeviceRequest("nvidia", 1, []string{"0"}, []string{"gpu", "compute", "utility"}),
	)

	// Request any two NVIDIA GPUs
	myContainer.SetHostOptions(
		hostoptions.DeviceRequest("nvidia", 2, nil, []string{"gpu", "compute"}),
	)
*/
func DeviceRequest(driver string, count int, deviceIDs []string, capabilities []string) SetHostOptFn {
	return func(opt *container.HostConfig) {
		if opt.DeviceRequests == nil {
			opt.DeviceRequests = make([]container.DeviceRequest, 0)
		}

		// Convert capabilities to the required format
		var caps [][]string
		if len(capabilities) > 0 {
			caps = [][]string{capabilities}
		}

		opt.DeviceRequests = append(opt.DeviceRequests, container.DeviceRequest{
			Driver:       driver,
			Count:        count,
			DeviceIDs:    deviceIDs,
			Capabilities: caps,
		})
	}
}
