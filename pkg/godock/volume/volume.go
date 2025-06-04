package volume

import (
	"github.com/aptd3v/godock/pkg/godock/volumeoptions"
	"github.com/docker/docker/api/types/volume"
)

// Volume represents a Docker volume along with its creation options.
type VolumeConfig struct {
	Options *volume.CreateOptions
}

// returns the volume name
func (v *VolumeConfig) String() string {
	return v.Options.Name
}

// SetOptions configures options for the Docker volume.
// Use this method to set various volume options using functions from the volumeopt package.
func (v *VolumeConfig) SetOptions(setVOFns ...volumeoptions.SetVolumeOptFn) {
	for _, set := range setVOFns {
		set(v.Options)
	}
}

// NewVolume creates a new Volume instance with the specified name.
// The Volume instance contains configuration options for creating a Docker volume.
func NewConfig(name string) *VolumeConfig {
	return &VolumeConfig{Options: &volume.CreateOptions{Name: name}}
}
