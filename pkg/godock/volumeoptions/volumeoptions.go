package volumeoptions

import (
	"github.com/docker/docker/api/types/volume"
)

// SetVolumeOptFn is a function type that configures options for creating a Docker volume.
type SetVolumeOptFn func(options *volume.CreateOptions)

// VolumeDriver represents supported volume drivers
type VolumeDriver string

const (
	// LocalDriver is the default local volume driver
	LocalDriver VolumeDriver = "local"
	// NFSDriver is for NFS volumes
	NFSDriver VolumeDriver = "nfs"
	// CephDriver is for Ceph RBD volumes
	CephDriver VolumeDriver = "rbd"
	// GlusterFSDriver is for GlusterFS volumes
	GlusterFSDriver VolumeDriver = "glusterfs"
)

/*
SetDriver sets the driver to use for creating the Docker volume.

Usage example:

	volume.SetOptions(
		volumeoptions.SetDriver(volumeoptions.NFSDriver),
	)
*/
func SetDriver(driver VolumeDriver) SetVolumeOptFn {
	return func(options *volume.CreateOptions) {
		options.Driver = string(driver)
	}
}

/*
AddDriverOpt adds a driver-specific option.
Each driver has its own set of options. Common ones include:

Local driver:
- size: Size of the volume
- type: Filesystem type

NFS driver:
- server: NFS server address
- share: Export path from NFS server
- security: NFS security options

Usage example:

	volume.SetOptions(
		volumeoptions.SetDriver(volumeoptions.NFSDriver),
		volumeoptions.AddDriverOpt("server", "10.0.0.1"),
		volumeoptions.AddDriverOpt("share", "/exports"),
	)
*/
func AddDriverOpt(key, value string) SetVolumeOptFn {
	return func(options *volume.CreateOptions) {
		if options.DriverOpts == nil {
			options.DriverOpts = make(map[string]string)
		}
		options.DriverOpts[key] = value
	}
}

/*
SetDriverOpts sets multiple driver-specific options at once.

Usage example:

	volume.SetOptions(
		volumeoptions.SetDriver(volumeoptions.NFSDriver),
		volumeoptions.SetDriverOpts(map[string]string{
			"server": "10.0.0.1",
			"share": "/exports",
			"security": "sys",
		}),
	)
*/
func SetDriverOpts(opts map[string]string) SetVolumeOptFn {
	return func(options *volume.CreateOptions) {
		if options.DriverOpts == nil {
			options.DriverOpts = make(map[string]string)
		}
		for k, v := range opts {
			options.DriverOpts[k] = v
		}
	}
}

/*
SetName sets the name of the Docker volume.

Usage example:

	volume.SetOptions(
		volumeoptions.SetName("my-data-volume"),
	)
*/
func SetName(name string) SetVolumeOptFn {
	return func(options *volume.CreateOptions) {
		options.Name = name
	}
}

/*
AddLabel adds a label to the Docker volume.
Labels provide a way to attach metadata to volumes for categorization or organization.

Usage example:

	volume.SetOptions(
		volumeoptions.AddLabel("environment", "production"),
		volumeoptions.AddLabel("project", "web-app"),
	)
*/
func AddLabel(key, value string) SetVolumeOptFn {
	return func(options *volume.CreateOptions) {
		if options.Labels == nil {
			options.Labels = make(map[string]string)
		}
		options.Labels[key] = value
	}
}

/*
SetLabels sets multiple labels at once.

Usage example:

	volume.SetOptions(
		volumeoptions.SetLabels(map[string]string{
			"environment": "production",
			"project": "web-app",
			"team": "backend",
		}),
	)
*/
func SetLabels(labels map[string]string) SetVolumeOptFn {
	return func(options *volume.CreateOptions) {
		if options.Labels == nil {
			options.Labels = make(map[string]string)
		}
		for k, v := range labels {
			options.Labels[k] = v
		}
	}
}

// AccessMode represents volume access modes
type AccessMode string

const (
	// SingleNode indicates the volume can only be accessed from a single node
	SingleNode AccessMode = "single"
	// MultiNode indicates the volume can be accessed from multiple nodes
	MultiNode AccessMode = "multi"
)

// SharingMode represents volume sharing modes
type SharingMode string

const (
	// None indicates no sharing
	None SharingMode = "none"
	// ReadOnly indicates read-only sharing
	ReadOnly SharingMode = "read-only"
	// ReadWrite indicates read-write sharing
	ReadWrite SharingMode = "read-write"
)

// internal mapping functions
func toDockerScope(access AccessMode) volume.Scope {
	switch access {
	case SingleNode:
		return "single"
	case MultiNode:
		return "multi"
	default:
		return "single"
	}
}

func toDockerSharingMode(sharing SharingMode) volume.SharingMode {
	switch sharing {
	case None:
		return "none"
	case ReadOnly:
		return "read-only"
	case ReadWrite:
		return "read-write"
	default:
		return "none"
	}
}

/*
SetClusterSpec sets the cluster volume specification for swarm mode volumes.

Usage example:

	volume.SetOptions(
		volumeoptions.SetClusterSpec(
			"backend-group",
			volumeoptions.SingleNode,
			volumeoptions.ReadWrite,
		),
	)
*/
func SetClusterSpec(group string, access AccessMode, sharing SharingMode) SetVolumeOptFn {
	return func(options *volume.CreateOptions) {
		options.ClusterVolumeSpec = &volume.ClusterVolumeSpec{
			Group: group,
			AccessMode: &volume.AccessMode{
				Scope:   toDockerScope(access),
				Sharing: toDockerSharingMode(sharing),
			},
		}
	}
}

/*
SetCapacityRange sets the desired capacity range for the volume.
This is only applicable for cluster volumes.

Usage example:

	volume.SetOptions(
		volumeoptions.SetCapacityRange(100*1024*1024, 1024*1024*1024), // 100MB to 1GB
	)
*/
func SetCapacityRange(requiredBytes, limitBytes int64) SetVolumeOptFn {
	return func(options *volume.CreateOptions) {
		if options.ClusterVolumeSpec == nil {
			options.ClusterVolumeSpec = &volume.ClusterVolumeSpec{}
		}
		options.ClusterVolumeSpec.CapacityRange = &volume.CapacityRange{
			RequiredBytes: requiredBytes,
			LimitBytes:    limitBytes,
		}
	}
}

// VolumeAvailability represents the availability state of a cluster volume
type VolumeAvailability = volume.Availability

const (
	// AvailabilityActive indicates that the volume is active and fully schedulable
	AvailabilityActive VolumeAvailability = volume.AvailabilityActive
	// AvailabilityPause indicates that no new workloads should use the volume
	AvailabilityPause VolumeAvailability = volume.AvailabilityPause
	// AvailabilityDrain indicates that all workloads using this volume should be rescheduled
	AvailabilityDrain VolumeAvailability = volume.AvailabilityDrain
)

/*
SetAvailability sets the availability state of a cluster volume.

Usage example:

	volume.SetOptions(
		volumeoptions.SetAvailability(volumeoptions.AvailabilityActive),
	)
*/
func SetAvailability(availability VolumeAvailability) SetVolumeOptFn {
	return func(options *volume.CreateOptions) {
		if options.ClusterVolumeSpec == nil {
			options.ClusterVolumeSpec = &volume.ClusterVolumeSpec{}
		}
		options.ClusterVolumeSpec.Availability = availability
	}
}

/*
AddSecret adds a secret that will be passed to the CSI storage plugin.
This is only applicable for cluster volumes.

Usage example:

	volume.SetOptions(
		volumeoptions.AddSecret("encryption-key", "my-secret-id"),
	)
*/
func AddSecret(key, secretID string) SetVolumeOptFn {
	return func(options *volume.CreateOptions) {
		if options.ClusterVolumeSpec == nil {
			options.ClusterVolumeSpec = &volume.ClusterVolumeSpec{}
		}
		options.ClusterVolumeSpec.Secrets = append(options.ClusterVolumeSpec.Secrets, volume.Secret{
			Key:    key,
			Secret: secretID,
		})
	}
}

// TopologyRequirement specifies where in the cluster a volume must be accessible from
type TopologyRequirement struct {
	Requisite []map[string]string
	Preferred []map[string]string
}

/*
SetTopologyRequirement sets the topology requirements for a cluster volume.
This specifies which nodes in the cluster the volume must be accessible from.

Usage example:

	volume.SetOptions(
		volumeoptions.SetTopologyRequirement(volumeoptions.TopologyRequirement{
			Requisite: []map[string]string{
				{"region": "us-east", "zone": "us-east-1a"},
				{"region": "us-east", "zone": "us-east-1b"},
			},
			Preferred: []map[string]string{
				{"region": "us-east", "zone": "us-east-1a"},
			},
		}),
	)
*/
func SetTopologyRequirement(req TopologyRequirement) SetVolumeOptFn {
	return func(options *volume.CreateOptions) {
		if options.ClusterVolumeSpec == nil {
			options.ClusterVolumeSpec = &volume.ClusterVolumeSpec{}
		}

		tr := &volume.TopologyRequirement{}

		for _, r := range req.Requisite {
			topology := volume.Topology{Segments: r}
			tr.Requisite = append(tr.Requisite, topology)
		}

		for _, p := range req.Preferred {
			topology := volume.Topology{Segments: p}
			tr.Preferred = append(tr.Preferred, topology)
		}

		options.ClusterVolumeSpec.AccessibilityRequirements = tr
	}
}
