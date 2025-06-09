package godock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/docker/docker/api/types/container"
)

type formater struct{ writer io.Writer }

func (f *formater) Write(p []byte) (n int, err error) {
	var data ContainerStats
	decoder := json.NewDecoder(bytes.NewReader(p))
	err = decoder.Decode(&data)
	if err != nil {
		return 0, err
	}
	b, err := json.Marshal(&FormatedContainerStats{
		CpuUsage:    data.FormatCpuUsagePercentage(),
		MemoryUsage: data.FormatMemoryUsage(),
		NetworkIO:   data.FormatNetworkIO(),
		DiskIO:      data.FormatDiskIO(),
	})
	if err != nil {
		return 0, err
	}
	if n, err = f.writer.Write(append(b, '\n')); err != nil {
		return n, err
	}

	return len(p), nil
}

// Formats the incoming stats and passes it to the supplied writer
func StatsFormatter(writer io.Writer) *formater {
	// container.Stats{}
	return &formater{writer: writer}
}

type FormatedContainerStats struct {
	CpuUsage    string `json:"cpuUsage"`
	MemoryUsage string `json:"memoryUsage"`
	NetworkIO   string `json:"networkIO"`
	DiskIO      string `json:"diskIO"`
}

type ContainerStats struct {
	Read         time.Time               `json:"read"`
	Preread      time.Time               `json:"preread"`
	PidsStats    container.PidsStats     `json:"pids_stats"`
	BlkioStats   container.BlkioStats    `json:"blkio_stats"`
	NumProcs     int64                   `json:"num_procs"`
	StorageStats container.StorageStats  `json:"storage_stats"`
	CpuStats     container.CPUStats      `json:"cpu_stats"`
	PreCPUStats  container.CPUStats      `json:"precpu_stats"`
	MemoryStats  container.MemoryStats   `json:"memory_stats"`
	Networks     map[string]NetworkStats `json:"networks"`
}

type NetworkStats struct {
	RxBytes   uint64 `json:"rx_bytes"`
	RxDropped uint64 `json:"rx_dropped"`
	RxErrors  uint64 `json:"rx_errors"`
	RxPackets uint64 `json:"rx_packets"`
	TxBytes   uint64 `json:"tx_bytes"`
	TxDropped uint64 `json:"tx_dropped"`
	TxErrors  uint64 `json:"tx_errors"`
	TxPackets uint64 `json:"tx_packets"`
}

func (stats *ContainerStats) FormatCpuUsagePercentage() string {
	// Calculate the total CPU time used by the container
	totalCPUUsage := float64(stats.CpuStats.CPUUsage.TotalUsage - stats.PreCPUStats.CPUUsage.TotalUsage)

	// Calculate the system CPU time
	systemCPUUsage := float64(stats.CpuStats.SystemUsage - stats.PreCPUStats.SystemUsage)

	// Calculate the number of online CPUs
	onlineCPUs := float64(stats.CpuStats.OnlineCPUs)

	// Calculate the CPU usage percentage
	cpuUsagePercentage := (totalCPUUsage / systemCPUUsage) * onlineCPUs * 100.0
	if math.IsNaN(cpuUsagePercentage) {
		return "0.00%"
	}
	return fmt.Sprintf("%.2f%%", cpuUsagePercentage)
}
func (stats *ContainerStats) FormatMemoryUsage() string {
	// Get the memory usage and limit in bytes
	memoryUsage := stats.MemoryStats.Usage
	memoryLimit := stats.MemoryStats.Limit

	// Convert the memory usage and limit to human-readable strings
	memoryUsageStr := bytesToHumanReadable(int64(memoryUsage))
	memoryLimitStr := bytesToHumanReadable(int64(memoryLimit))

	// Combine the strings and return the result
	return fmt.Sprintf("%s / %s", memoryUsageStr, memoryLimitStr)
}

func (stats *ContainerStats) FormatDiskIO() string {
	// Get the disk read/write values in bytes
	var readBytes, writeBytes uint64

	// Sum up all read and write operations
	for _, stat := range stats.BlkioStats.IoServiceBytesRecursive {
		switch stat.Op {
		case "Read":
			readBytes += stat.Value
		case "Write":
			writeBytes += stat.Value
		}
	}

	// Convert the disk read/write values to human-readable strings
	readBytesStr := bytesToHumanReadable(int64(readBytes))
	writeBytesStr := bytesToHumanReadable(int64(writeBytes))

	// Combine the strings and return the result
	return fmt.Sprintf("%s / %s", readBytesStr, writeBytesStr)
}

func (stats *ContainerStats) FormatNetworkIO() string {
	var totalRx, totalTx uint64

	// Sum up all network interfaces
	for _, net := range stats.Networks {
		totalRx += net.RxBytes
		totalTx += net.TxBytes
	}

	// Convert to human readable format
	rxStr := bytesToHumanReadable(int64(totalRx))
	txStr := bytesToHumanReadable(int64(totalTx))

	return fmt.Sprintf("%s / %s", rxStr, txStr)
}

func bytesToHumanReadable(bytes int64) string {
	// Define the units and their corresponding values in bytes
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}

	// If the value is already in bytes or less than 1KB, return it directly
	if bytes < 1024 {
		return fmt.Sprintf("%d %s", bytes, units[0])
	}

	// Calculate the index to get the appropriate unit from the "units" array
	index := 0
	value := float64(bytes)
	for value >= 1024 && index < len(units)-1 {
		value /= 1024
		index++
	}

	// Format the value with 2 decimal places and return the human-readable string
	return fmt.Sprintf("%.2f %s", value, units[index])
}
