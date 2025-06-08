package imageoptions

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"runtime"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
)

// SetPullOptFn is a function type that configures pull options for a Docker image.
type SetPullOptFn func(options *image.PullOptions)

// SetBuildOptFn is a function type that configures build options for a Docker image.
type SetBuildOptFn func(options *types.ImageBuildOptions)

// SetPushOptFn is a function type that configures push options for a Docker image.
type SetPushOptFn func(options *image.PushOptions)

// BuilderVersion represents the version of the builder to use
type BuilderVersion string

const (
	// BuilderV1 uses the legacy builder
	BuilderV1 BuilderVersion = "1"
	// BuilderV2 uses BuildKit
	BuilderV2 BuilderVersion = "2"
)

// OutputType represents the type of build output
type OutputType string

const (
	// LocalOutput represents output to a local directory
	LocalOutput OutputType = "local"
	// TarOutput represents output to a tar file
	TarOutput OutputType = "tar"
	// ImageOutput represents output as a Docker image
	ImageOutput OutputType = "image"
)

// Auth represents registry authentication credentials
type Auth struct {
	Username string
	Password string
}

// encodeAuthToBase64 converts auth credentials to base64 encoded auth string
func encodeAuthToBase64(auth Auth) string {
	authConfig := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: auth.Username,
		Password: auth.Password,
	}

	jsonAuth, err := json.Marshal(authConfig)
	if err != nil {
		return ""
	}

	return base64.URLEncoding.EncodeToString(jsonAuth)
}

/*
RemoveIntermediateContainers specifies whether to remove intermediate containers after a successful build.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.RemoveIntermediateContainers(true),
	)
*/
func RemoveIntermediateContainers(remove bool) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.Remove = remove
	}
}

/*
AddTag adds a tag to the image.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.AddTag("v1.0.0"),
		imageoptions.AddTag("latest"),
	)
*/
func AddTag(tag string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		if options.Tags == nil {
			options.Tags = make([]string, 0)
		}
		options.Tags = append(options.Tags, tag)
	}
}

/*
CacheEnabled controls whether to use the build cache.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.CacheEnabled(false), // Disable cache
	)
*/
func CacheEnabled(enabled bool) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.NoCache = !enabled
	}
}

/*
SetDockerfile specifies the path to the Dockerfile.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetDockerfile("./Dockerfile.prod"),
	)
*/
func SetDockerfile(path string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.Dockerfile = path
	}
}

/*
SetBuildContext provides the build context for the image.

Usage example:

	tarFile := createTarContext() // Your tar creation logic
	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetBuildContext(bytes.NewReader(tarFile)),
	)
*/
func SetBuildContext(context io.Reader) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.Context = context
	}
}

/*
SetBuildArgs sets build-time variables.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetBuildArgs(map[string]*string{
			"VERSION": &version,
			"DEBUG": &debug,
		}),
	)
*/
func SetBuildArgs(args map[string]*string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.BuildArgs = args
	}
}

/*
SetTarget sets the target build stage to build.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetTarget("prod"),
	)
*/
func SetTarget(target string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.Target = target
	}
}

/*
SetNetworkMode sets the networking mode for the build.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetNetworkMode("host"),
	)
*/
func SetNetworkMode(networkMode string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.NetworkMode = networkMode
	}
}

/*
SetCgroupParent sets the parent cgroup for build containers.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetCgroupParent("/custom/cgroup"),
	)
*/
func SetCgroupParent(cgroupParent string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.CgroupParent = cgroupParent
	}
}

/*
AllowExtraHosts adds entries to /etc/hosts in building containers.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.AllowExtraHosts([]string{
			"myhost:10.10.10.10",
			"myother:10.10.10.11",
		}),
	)
*/
func AllowExtraHosts(extraHosts []string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.ExtraHosts = extraHosts
	}
}

/*
SetIsolation sets container isolation technology.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetIsolation("hyperv"),
	)
*/
func SetIsolation(isolation string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.Isolation = container.Isolation(isolation)
	}
}

/*
SetMemory sets memory limit for build containers.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetMemory(1024*1024*1024), // 1GB
	)
*/
func SetMemory(memory int64) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.Memory = memory
	}
}

/*
SetCPUSetCPUs sets which CPUs to use for build containers.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetCPUSetCPUs("0-3"),
	)
*/
func SetCPUSetCPUs(cpusetcpus string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.CPUSetCPUs = cpusetcpus
	}
}

/*
SetCPUShares sets CPU shares for build containers (relative weight).

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetCPUShares(1024),
	)
*/
func SetCPUShares(cpushares int64) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.CPUShares = cpushares
	}
}

/*
SetSecurityOpt sets security options for build containers.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetSecurityOpt([]string{
			"seccomp=unconfined",
			"apparmor=unconfined",
		}),
	)
*/
func SetSecurityOpt(securityOpt []string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.SecurityOpt = securityOpt
	}
}

/*
SetShmSize sets the size of /dev/shm for build containers.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetShmSize(67108864), // 64MB
	)
*/
func SetShmSize(shmSize int64) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.ShmSize = shmSize
	}
}

/*
SetUlimitNoFile sets the file descriptor limits for build containers.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetUlimitNoFile(65535, 65535), // soft limit, hard limit
	)
*/
func SetUlimitNoFile(soft, hard int64) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		if options.Ulimits == nil {
			options.Ulimits = make([]*container.Ulimit, 0)
		}
		// Replace any existing nofile ulimit
		for i, u := range options.Ulimits {
			if u.Name == "nofile" {
				options.Ulimits[i] = &container.Ulimit{
					Name: "nofile",
					Soft: soft,
					Hard: hard,
				}
				return
			}
		}
		// Add new nofile ulimit if not found
		options.Ulimits = append(options.Ulimits, &container.Ulimit{
			Name: "nofile",
			Soft: soft,
			Hard: hard,
		})
	}
}

/*
SetUlimitNProc sets the process limit for build containers.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetUlimitNProc(4096, 8192), // soft limit, hard limit
	)
*/
func SetUlimitNProc(soft, hard int64) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		if options.Ulimits == nil {
			options.Ulimits = make([]*container.Ulimit, 0)
		}
		// Replace any existing nproc ulimit
		for i, u := range options.Ulimits {
			if u.Name == "nproc" {
				options.Ulimits[i] = &container.Ulimit{
					Name: "nproc",
					Soft: soft,
					Hard: hard,
				}
				return
			}
		}
		// Add new nproc ulimit if not found
		options.Ulimits = append(options.Ulimits, &container.Ulimit{
			Name: "nproc",
			Soft: soft,
			Hard: hard,
		})
	}
}

/*
SetUlimitCore sets the core file size limit for build containers.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetUlimitCore(0, 0), // disable core dumps
	)
*/
func SetUlimitCore(soft, hard int64) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		if options.Ulimits == nil {
			options.Ulimits = make([]*container.Ulimit, 0)
		}
		// Replace any existing core ulimit
		for i, u := range options.Ulimits {
			if u.Name == "core" {
				options.Ulimits[i] = &container.Ulimit{
					Name: "core",
					Soft: soft,
					Hard: hard,
				}
				return
			}
		}
		// Add new core ulimit if not found
		options.Ulimits = append(options.Ulimits, &container.Ulimit{
			Name: "core",
			Soft: soft,
			Hard: hard,
		})
	}
}

/*
SetUlimitMemlock sets the locked-in-memory address space limit for build containers.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetUlimitMemlock(65536, 65536), // 64KB soft and hard limit
	)
*/
func SetUlimitMemlock(soft, hard int64) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		if options.Ulimits == nil {
			options.Ulimits = make([]*container.Ulimit, 0)
		}
		// Replace any existing memlock ulimit
		for i, u := range options.Ulimits {
			if u.Name == "memlock" {
				options.Ulimits[i] = &container.Ulimit{
					Name: "memlock",
					Soft: soft,
					Hard: hard,
				}
				return
			}
		}
		// Add new memlock ulimit if not found
		options.Ulimits = append(options.Ulimits, &container.Ulimit{
			Name: "memlock",
			Soft: soft,
			Hard: hard,
		})
	}
}

/*
SetUlimitRtPrio sets the real-time priority limit for build containers.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetUlimitRtPrio(0, 0), // disable real-time priority
	)
*/
func SetUlimitRtPrio(soft, hard int64) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		if options.Ulimits == nil {
			options.Ulimits = make([]*container.Ulimit, 0)
		}
		// Replace any existing rtprio ulimit
		for i, u := range options.Ulimits {
			if u.Name == "rtprio" {
				options.Ulimits[i] = &container.Ulimit{
					Name: "rtprio",
					Soft: soft,
					Hard: hard,
				}
				return
			}
		}
		// Add new rtprio ulimit if not found
		options.Ulimits = append(options.Ulimits, &container.Ulimit{
			Name: "rtprio",
			Soft: soft,
			Hard: hard,
		})
	}
}

/*
ForceRemove forces the removal of the image even if it is being used by stopped containers.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.ForceRemove(true),
	)
*/
func ForceRemove(force bool) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.ForceRemove = force
	}
}

/*
SetSquash squashes newly built layers into a single new layer.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetSquash(true),
	)
*/
func SetSquash(squash bool) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.Squash = squash
	}
}

/*
AddLabel adds a label to the built image.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.AddLabel("org.label-schema.version", "1.0.0"),
	)
*/
func AddLabel(key, value string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		if options.Labels == nil {
			options.Labels = make(map[string]string)
		}
		options.Labels[key] = value
	}
}

/*
SetPullParent controls whether to pull the parent image.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetPullParent(true),
	)
*/
func SetPullParent(pull bool) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.PullParent = pull
	}
}

/*
SetBuilderVersion sets the version of the builder to use.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetBuilderVersion(imageoptions.BuilderV2),
	)
*/
func SetBuilderVersion(version BuilderVersion) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.Version = types.BuilderVersion(version)
	}
}

/*
SetBuildID sets a unique ID for the build.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetBuildID("build-123"),
	)
*/
func SetBuildID(buildID string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.BuildID = buildID
	}
}

/*
SetSessionID sets a unique session ID for the build.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetSessionID("session-123"),
	)
*/
func SetSessionID(sessionID string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.SessionID = sessionID
	}
}

/*
SetBuildInfo sets build info arguments for the build.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetBuildInfo(map[string]string{
			"key": "value",
		}),
	)
*/
func SetBuildInfo(args map[string]string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		if options.BuildArgs == nil {
			options.BuildArgs = make(map[string]*string)
		}
		for k, v := range args {
			val := v // Create a new variable to hold the value
			options.BuildArgs[k] = &val
		}
	}
}

/*
SetBuildPlatform sets the platform for the build.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetBuildPlatform("linux/amd64"),
	)
*/
func SetBuildPlatform(platform string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.Platform = platform
	}
}

/*
SetPullPolicy sets the pull policy for the build.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetPullPolicy("always"),
	)
*/
func SetPullPolicy(policy string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.PullParent = policy == "always"
	}
}

/*
SetExtraHosts adds extra hosts to /etc/hosts in building containers.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		imageoptions.SetExtraHosts([]string{
			"myhost:10.10.10.10",
			"myother:10.10.10.11",
		}),
	)
*/
func SetExtraHosts(hosts []string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.ExtraHosts = hosts
	}
}

// PULL OPTIONS

/*
PullAllTags specifies whether to download all tagged images in the repository.

Usage example:

	img := image.NewConfig("alpine")
	img.SetPullOptions(
		imageoptions.PullAllTags(true),
	)
*/
func PullAllTags(enabled bool) SetPullOptFn {
	return func(options *image.PullOptions) {
		options.All = enabled
	}
}

/*
SetPullPlatform sets the platform for image pulling.

Usage example:

	img := image.NewConfig("my-image")
	img.SetPullOptions(
		imageoptions.SetPullPlatform("linux/amd64"),
	)
*/
func SetPullPlatform(platform string) SetPullOptFn {
	return func(options *image.PullOptions) {
		options.Platform = platform
	}
}

/*
UseCurrentPlatform sets the platform to match the current system architecture.

Usage example:

	img := image.NewConfig("alpine")
	img.SetPullOptions(
		imageoptions.UseCurrentPlatform(),
	)
*/
func UseCurrentPlatform() SetPullOptFn {
	return func(options *image.PullOptions) {
		options.Platform = runtime.GOARCH
	}
}

/*
SetPrivilegeFunc sets the authentication function for restricted access images.

Usage example:

	img := image.NewConfig("private-repo/my-image")
	img.SetPullOptions(
		imageoptions.SetPrivilegeFunc(func(ctx context.Context) (string, error) {
			return "Bearer " + os.Getenv("DOCKER_TOKEN"), nil
		}),
	)
*/
func SetPrivilegeFunc(authFn func(ctx context.Context) (string, error)) SetPullOptFn {
	return func(options *image.PullOptions) {
		options.PrivilegeFunc = authFn
	}
}

/*
SetRegistryAuth sets the registry authentication credentials for pushing or pulling images.

Usage example for pull:

	image.SetPullOptions(
		imageoptions.SetRegistryAuth("username", "password"),
	)

Usage example for push:

	image.SetPushOptions(
		imageoptions.SetRegistryAuth("username", "password"),
	)
*/
func SetRegistryAuth(username, password string) interface{} {
	auth := Auth{
		Username: username,
		Password: password,
	}
	encodedAuth := encodeAuthToBase64(auth)

	return struct {
		SetPullOptFn
		SetPushOptFn
	}{
		SetPullOptFn: func(options *image.PullOptions) {
			options.RegistryAuth = encodedAuth
		},
		SetPushOptFn: func(options *image.PushOptions) {
			options.RegistryAuth = encodedAuth
		},
	}
}

/*
SetPlatform sets the platform for pulling multi-platform images.

Usage example:

	image.SetPullOptions(
		imageoptions.SetPlatform("linux/amd64"),
	)
*/
func SetPlatform(platform string) SetPullOptFn {
	return func(options *image.PullOptions) {
		options.Platform = platform
	}
}

/*
AddOutput adds an output configuration for the build.

Usage example:

	img := image.NewConfig("my-image")
	img.SetBuildOptions(
		// Output to a local directory
		imageoptions.AddOutput(imageoptions.LocalOutput, map[string]string{
			"dest": "/tmp/output",
		}),
		// Output as a tar file
		imageoptions.AddOutput(imageoptions.TarOutput, map[string]string{
			"dest": "/tmp/output.tar",
		}),
		// Output as an image
		imageoptions.AddOutput(imageoptions.ImageOutput, map[string]string{
			"name": "my-org/my-image:latest",
		}),
	)
*/
func AddOutput(outputType OutputType, attrs map[string]string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		if options.Outputs == nil {
			options.Outputs = make([]types.ImageBuildOutput, 0)
		}
		options.Outputs = append(options.Outputs, types.ImageBuildOutput{
			Type:  string(outputType),
			Attrs: attrs,
		})
	}
}
