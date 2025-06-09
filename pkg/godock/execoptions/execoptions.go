package execoptions

import (
	"fmt"

	containerType "github.com/docker/docker/api/types/container"
)

type ExecOptionsFn func(options *containerType.ExecOptions)

func TTY(tty bool) ExecOptionsFn {
	return func(options *containerType.ExecOptions) {
		options.Tty = tty
	}
}
func AttachStdin(attachStdin bool) ExecOptionsFn {
	return func(options *containerType.ExecOptions) {
		options.AttachStdin = attachStdin
	}
}
func AttachStdout(attachStdout bool) ExecOptionsFn {
	return func(options *containerType.ExecOptions) {
		options.AttachStdout = attachStdout
	}
}
func AttachStderr(attachStderr bool) ExecOptionsFn {
	return func(options *containerType.ExecOptions) {
		options.AttachStderr = attachStderr
	}
}
func Detach(detach bool) ExecOptionsFn {
	return func(options *containerType.ExecOptions) {
		options.Detach = detach
	}
}
func CMD(cmd ...string) ExecOptionsFn {
	return func(options *containerType.ExecOptions) {
		if options.Cmd == nil {
			options.Cmd = []string{}
		}
		options.Cmd = append(options.Cmd, cmd...)
	}
}
func User(user string) ExecOptionsFn {
	return func(options *containerType.ExecOptions) {
		options.User = user
	}
}
func ENV(key, value string) ExecOptionsFn {
	return func(options *containerType.ExecOptions) {
		if options.Env == nil {
			options.Env = []string{}
		}
		options.Env = append(options.Env, fmt.Sprintf("%s=%s", key, value))
	}
}
func WorkingDir(workingDir string) ExecOptionsFn {
	return func(options *containerType.ExecOptions) {
		options.WorkingDir = workingDir
	}
}
func Privileged(privileged bool) ExecOptionsFn {
	return func(options *containerType.ExecOptions) {
		options.Privileged = privileged
	}
}
func DetachKeys(detachKeys string) ExecOptionsFn {
	return func(options *containerType.ExecOptions) {
		options.DetachKeys = detachKeys
	}
}
func ConsoleSize(width, height uint) ExecOptionsFn {
	return func(options *containerType.ExecOptions) {
		if options.ConsoleSize == nil {
			options.ConsoleSize = &[2]uint{0, 0}
		}
		options.ConsoleSize = &[2]uint{width, height}
	}
}
