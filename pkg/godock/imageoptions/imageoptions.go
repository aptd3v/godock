package imageoptions

import (
	"context"
	"encoding/base64"
	"io"
	"runtime"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
)

type SetBuildOptFn func(options *types.ImageBuildOptions)

/*
Specify to remove intermediate containers after a successful build

	img := image.NewConfig("node:latest")
	img.S(
		Remove(),
	)
*/
func Remove() SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.Remove = true
	}
}

/*
Specify not to remove intermediate containers after a successful build

	img := image.NewConfig("node:latest")
	img.S(
		DontRemove(),
	)
*/
func DontRemove() SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.Remove = false
	}
}

/*
Adds a tag to the image reference

	img := image.NewConfig("node:latest")
	img.S(
		Tag("node:latest"),
	)
*/
func Tag(tag string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		if options.Tags == nil {
			options.Tags = make([]string, 0)
		}
		options.Tags = append(options.Tags, tag)
	}
}

/*
Sets the Dockerfile location for the image

	img := image.NewConfig("node:latest")
	img.S(
		DisableCache()
	)
*/
func DisableCache() SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.NoCache = true
	}
}

/*
Sets the Dockerfile location for the image

	img := image.NewConfig("node:latest")
	img.SetBuildOptions(
		Dockerfile("/app/Dockerfile")
	)
*/
func Dockerfile(path string) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.Dockerfile = path
	}
}

/*
Sets the build context for the image

	img := image.NewConfig("node:latest")
	img.S(
		BuildContext(bytes.NewReader(tarFile)),
	)
*/
func BuildContext(context io.Reader) SetBuildOptFn {
	return func(options *types.ImageBuildOptions) {
		options.Context = context
	}
}

// PULL OPTIONS

type SetPullOptFn func(options *image.PullOptions)

/*
Download all tagged images in the repository.

Short hand equivalent:
"--all-tags , -a"

	img := image.NewConfig("alpine")
	img.PullOptions(
		pulloptions.DownloadAllTaggedImages()
	)
*/
func DownloadAllTaggedImages() SetPullOptFn {
	return func(options *image.PullOptions) {
		options.All = true
	}
}

/*
Sets platform if server is multi-platform capable

Short hand equivalent:
"--platform"

	img := image.NewConfig("alpine")
	img.PullOptions(
		pulloptions.Platform("linux/arm64")
	)
*/
func Platform(platform string) SetPullOptFn {
	return func(options *image.PullOptions) {
		options.Platform = platform
	}
}

// Matches the platform to the runtime's arch
func MatchPlatform() SetPullOptFn {
	return func(options *image.PullOptions) {
		options.Platform = runtime.GOARCH
	}
}

/*
Sets the privilege function sed to
authenticate or authorize the pull
operation for images that have restricted access.

	img := image.NewConfig("alpine")
	img.PullOptions(
		pulloptions.Privilege(func(ctx) (string, error) {
			// Perform your authentication logic here
			// For example, return an authentication token
			return "Bearer <your-auth-token>", nil
		})
	)
*/
func Privilege(pFn func(ctx context.Context) (string, error)) SetPullOptFn {
	return func(options *image.PullOptions) {
		options.PrivilegeFunc = pFn
	}
}

/*
Sets the base64 encoded credentials for the registry

	img := image.NewConfig("alpine")
	img.PullOptions(
		pulloptions.RegistryAuth("user", "pass")
	)
*/
func RegistryAuth(username, password string) SetPullOptFn {
	encoded := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return func(options *image.PullOptions) {
		options.RegistryAuth = encoded
	}
}
