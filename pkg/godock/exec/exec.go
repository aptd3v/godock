package exec

import (
	"github.com/aptd3v/godock/pkg/godock/execoptions"
	containerType "github.com/docker/docker/api/types/container"
)

type ExecConfig struct {
	Options *containerType.ExecOptions
	ID      string
}

func NewConfig() *ExecConfig {
	return &ExecConfig{
		Options: &containerType.ExecOptions{},
	}
}

func (c *ExecConfig) SetOptions(options ...execoptions.ExecOptionsFn) {
	for _, option := range options {
		option(c.Options)
	}
}

// SetUser sets the user that will run the command
func (c *ExecConfig) SetUser(user string) *ExecConfig {
	c.Options.User = user
	return c
}

// SetPrivileged sets whether the container is in privileged mode
func (c *ExecConfig) SetPrivileged(privileged bool) *ExecConfig {
	c.Options.Privileged = privileged
	return c
}

// SetTty sets whether to attach standard streams to a tty
func (c *ExecConfig) SetTty(tty bool) *ExecConfig {
	c.Options.Tty = tty
	return c
}

// SetConsoleSize sets the initial console size [height, width]
func (c *ExecConfig) SetConsoleSize(height, width uint) *ExecConfig {
	c.Options.ConsoleSize = &[2]uint{height, width}
	return c
}

// SetAttachStdin sets whether to attach the standard input
func (c *ExecConfig) SetAttachStdin(attach bool) *ExecConfig {
	c.Options.AttachStdin = attach
	return c
}

// SetAttachStderr sets whether to attach the standard error
func (c *ExecConfig) SetAttachStderr(attach bool) *ExecConfig {
	c.Options.AttachStderr = attach
	return c
}

// SetAttachStdout sets whether to attach the standard output
func (c *ExecConfig) SetAttachStdout(attach bool) *ExecConfig {
	c.Options.AttachStdout = attach
	return c
}

// SetDetach sets whether to execute in detach mode
func (c *ExecConfig) SetDetach(detach bool) *ExecConfig {
	c.Options.Detach = detach
	return c
}

// SetDetachKeys sets the escape keys for detach
func (c *ExecConfig) SetDetachKeys(keys string) *ExecConfig {
	c.Options.DetachKeys = keys
	return c
}

// SetEnv sets multiple environment variables
func (c *ExecConfig) SetEnv(env []string) *ExecConfig {
	c.Options.Env = env
	return c
}

// AddEnv adds a single environment variable
func (c *ExecConfig) AddEnv(key, value string) *ExecConfig {
	if c.Options.Env == nil {
		c.Options.Env = make([]string, 0)
	}
	c.Options.Env = append(c.Options.Env, key+"="+value)
	return c
}

// SetWorkingDir sets the working directory
func (c *ExecConfig) SetWorkingDir(dir string) *ExecConfig {
	c.Options.WorkingDir = dir
	return c
}

// SetCmd sets the execution commands and args
func (c *ExecConfig) SetCmd(cmd ...string) *ExecConfig {
	c.Options.Cmd = cmd
	return c
}

// AddCmd adds commands and args to the existing command slice
func (c *ExecConfig) AddCmd(cmd ...string) *ExecConfig {
	if c.Options.Cmd == nil {
		c.Options.Cmd = make([]string, 0)
	}
	c.Options.Cmd = append(c.Options.Cmd, cmd...)
	return c
}
