package updateoptions

import (
	"github.com/docker/docker/api/types/blkiodev"
	containerType "github.com/docker/docker/api/types/container"

	"github.com/aptd3v/godock/pkg/godock"
)

// WithCPUShares updates the CPU shares for the container.
func WithCPUShares(cpuShares int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.CPUShares = cpuShares
	}
}

// WithMemory updates the memory for the container.
func WithMemory(memory int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.Memory = memory
	}
}

// WithNanoCPUs updates the nano CPUs for the container.
func WithNanoCPUs(nanoCPUs int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.NanoCPUs = nanoCPUs
	}
}

// WithCgroupParent updates the cgroup parent for the container.
func WithCgroupParent(cgroupParent string) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.CgroupParent = cgroupParent
	}
}

// WithBlkioWeight updates the blkio weight for the container.
func WithBlkioWeight(blkioWeight uint16) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.BlkioWeight = blkioWeight
	}
}

// WithBlkioWeightDevice updates the blkio weight device for the container.
func WithBlkioWeightDevice(path string, weight uint16) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		if options.BlkioWeightDevice == nil {
			options.BlkioWeightDevice = make([]*blkiodev.WeightDevice, 0)
		}
		options.BlkioWeightDevice = append(options.BlkioWeightDevice, &blkiodev.WeightDevice{
			Path:   path,
			Weight: weight,
		})
	}
}

// WithBlkioDeviceReadBps updates the blkio device read bps for the container.
func WithBlkioDeviceReadBps(path string, rate uint64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		if options.BlkioDeviceReadBps == nil {
			options.BlkioDeviceReadBps = make([]*blkiodev.ThrottleDevice, 0)
		}
		options.BlkioDeviceReadBps = append(options.BlkioDeviceReadBps, &blkiodev.ThrottleDevice{
			Path: path,
			Rate: rate,
		})
	}
}

// WithBlkioDeviceWriteBps updates the blkio device write bps for the container.
func WithBlkioDeviceWriteBps(path string, rate uint64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		if options.BlkioDeviceWriteBps == nil {
			options.BlkioDeviceWriteBps = make([]*blkiodev.ThrottleDevice, 0)
		}
	}
}

// WithBlkioDeviceReadIOps updates the blkio device read iops for the container.
func WithBlkioDeviceReadIOps(path string, rate uint64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		if options.BlkioDeviceReadIOps == nil {
			options.BlkioDeviceReadIOps = make([]*blkiodev.ThrottleDevice, 0)
		}
	}
}

// WithBlkioDeviceWriteIOps updates the blkio device write iops for the container.
func WithBlkioDeviceWriteIOps(path string, rate uint64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		if options.BlkioDeviceWriteIOps == nil {
			options.BlkioDeviceWriteIOps = make([]*blkiodev.ThrottleDevice, 0)
		}
	}
}

// WithCPUPeriod updates the cpu period for the container.
func WithCPUPeriod(cpuPeriod int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.CPUPeriod = cpuPeriod
	}
}

// WithCPUQuota updates the cpu quota for the container.
func WithCPUQuota(cpuQuota int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.CPUQuota = cpuQuota
	}
}

// WithCPURealtimePeriod updates the cpu realtime period for the container.
func WithCPURealtimePeriod(cpuRealtimePeriod int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.CPURealtimePeriod = cpuRealtimePeriod
	}
}

// WithCPURealtimeRuntime updates the cpu realtime runtime for the container.
func WithCPURealtimeRuntime(cpuRealtimeRuntime int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.CPURealtimeRuntime = cpuRealtimeRuntime
	}
}

// WithCpusetCpus updates the cpuset cpus for the container.
func WithCpusetCpus(cpusetCpus string) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.CpusetCpus = cpusetCpus
	}
}

// WithCpusetMems updates the cpuset mems for the container.
func WithCpusetMems(cpusetMems string) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.CpusetMems = cpusetMems
	}
}

// WithDevices updates the devices for the container.
func WithDevices(pathOnHost string, pathInContainer string, cgroupPermissions string) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		if options.Devices == nil {
			options.Devices = make([]containerType.DeviceMapping, 0)
		}
		options.Devices = append(options.Devices, containerType.DeviceMapping{
			PathOnHost:        pathOnHost,
			PathInContainer:   pathInContainer,
			CgroupPermissions: cgroupPermissions,
		})
	}
}

// WithDeviceCgroupRules updates the device cgroup rules for the container.
func WithDeviceCgroupRules(rules ...string) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		if options.DeviceCgroupRules == nil {
			options.DeviceCgroupRules = make([]string, 0)
		}
		options.DeviceCgroupRules = append(options.DeviceCgroupRules, rules...)
	}
}

// WithDeviceRequests updates the device requests for the container.
func WithDeviceRequests(driver string, count int, capabilities [][]string) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		if options.DeviceRequests == nil {
			options.DeviceRequests = make([]containerType.DeviceRequest, 0)
		}
		options.DeviceRequests = append(options.DeviceRequests, containerType.DeviceRequest{
			Driver:       driver,
			Count:        count,
			Capabilities: capabilities,
		})
	}
}

// WithKernelMemory updates the kernel memory for the container.
func WithKernelMemory(kernelMemory int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.KernelMemory = kernelMemory
	}
}

// WithKernelMemoryTCP updates the kernel memory tcp for the container.
func WithKernelMemoryTCP(kernelMemoryTCP int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.KernelMemoryTCP = kernelMemoryTCP
	}
}

// WithMemoryReservation updates the memory reservation for the container.
func WithMemoryReservation(memoryReservation int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.MemoryReservation = memoryReservation
	}
}

// WithMemorySwap updates the memory swap for the container.
func WithMemorySwap(memorySwap int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.MemorySwap = memorySwap
	}
}

// WithMemorySwappiness updates the memory swappiness for the container.
func WithMemorySwappiness(memorySwappiness int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.MemorySwappiness = &memorySwappiness
	}
}

// WithOomKillDisable updates the oom kill disable for the container.
func WithOomKillDisable(oomKillDisable bool) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.OomKillDisable = &oomKillDisable
	}
}

// WithPidsLimit updates the pids limit for the container.
func WithPidsLimit(pidsLimit int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.PidsLimit = &pidsLimit
	}
}

// WithUlimits updates the ulimits for the container.
func WithUlimits(name string, softLimit int64, hardLimit int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.Ulimits = append(options.Ulimits, &containerType.Ulimit{
			Name: name,
			Soft: softLimit,
			Hard: hardLimit,
		})
	}
}

// WithCPUCount updates the cpu count for the container.
func WithCPUCount(cpuCount int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.CPUCount = cpuCount
	}
}

// WithCPUPercent updates the cpu percent for the container.
func WithCPUPercent(cpuPercent int64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.CPUPercent = cpuPercent
	}
}

// WithIOMaximumIOps updates the io maximum iops for the container.
func WithIOMaximumIOps(ioMaximumIOps uint64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.IOMaximumIOps = ioMaximumIOps
	}
}

// WithIOMaximumBandwidth updates the io maximum bandwidth for the container.
func WithIOMaximumBandwidth(ioMaximumBandwidth uint64) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.IOMaximumBandwidth = ioMaximumBandwidth
	}
}

func WithRestartPolicy(name string, maximumRetryCount int) godock.UpdateOptionFn {
	return func(options *containerType.UpdateConfig) {
		options.RestartPolicy = containerType.RestartPolicy{
			Name:              containerType.RestartPolicyMode(name),
			MaximumRetryCount: maximumRetryCount,
		}
	}
}
