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
