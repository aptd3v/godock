package godock

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatsFormatter(t *testing.T) {
	var buf bytes.Buffer
	formatter := StatsFormatter(&buf)

	// Create sample stats data
	stats := ContainerStats{
		ID:   "test-container",
		Name: "test",
		CpuStats: struct {
			CpuUsage struct {
				TotalUsage        int64 `json:"totalUsage"`
				UsageInKernelMode int64 `json:"usageInKernelmode"`
				UsageInUserMode   int64 `json:"usageInUsermode"`
			} `json:"cpuUsage"`
			SystemCPUUsage int64 `json:"systemCpuUsage"`
			OnlineCPUs     int64 `json:"onlineCpus"`
			ThrottlingData struct {
				Periods          int64 `json:"periods"`
				ThrottledPeriods int64 `json:"throttledPeriods"`
				ThrottledTime    int64 `json:"throttledTime"`
			} `json:"throttlingData"`
		}{
			CpuUsage: struct {
				TotalUsage        int64 `json:"totalUsage"`
				UsageInKernelMode int64 `json:"usageInKernelmode"`
				UsageInUserMode   int64 `json:"usageInUsermode"`
			}{
				TotalUsage: 100000000,
			},
			SystemCPUUsage: 1000000000,
			OnlineCPUs:     4,
		},
		PreCPUStats: struct {
			CpuUsage struct {
				TotalUsage        int64 `json:"totalUsage"`
				UsageInKernelMode int64 `json:"usageInKernelmode"`
				UsageInUserMode   int64 `json:"usageInUsermode"`
			} `json:"cpuUsage"`
			SystemCPUUsage int64 `json:"systemCpuUsage"`
			OnlineCPUs     int64 `json:"onlineCpus"`
			ThrottlingData struct {
				Periods          int64 `json:"periods"`
				ThrottledPeriods int64 `json:"throttledPeriods"`
				ThrottledTime    int64 `json:"throttledTime"`
			} `json:"throttlingData"`
		}{
			CpuUsage: struct {
				TotalUsage        int64 `json:"totalUsage"`
				UsageInKernelMode int64 `json:"usageInKernelmode"`
				UsageInUserMode   int64 `json:"usageInUsermode"`
			}{
				TotalUsage: 0,
			},
			SystemCPUUsage: 0,
		},
		MemoryStats: struct {
			Usage int64 `json:"usage"`
			Stats struct {
				ActiveAnon          int64 `json:"activeAnon"`
				ActiveFile          int64 `json:"activeFile"`
				Anon                int64 `json:"anon"`
				AnonTHP             int64 `json:"anonTHP"`
				File                int64 `json:"file"`
				FileDirty           int64 `json:"fileDirty"`
				FileMapped          int64 `json:"fileMapped"`
				FileWriteBack       int64 `json:"fileWriteback"`
				InactiveAnon        int64 `json:"inactiveAnon"`
				InactiveFile        int64 `json:"inactiveFile"`
				KernelStack         int64 `json:"kernelStack"`
				PgActivate          int64 `json:"pgActivate"`
				PgDeactivate        int64 `json:"pgDeactivate"`
				PgFault             int64 `json:"pgFault"`
				PgLazyFree          int64 `json:"pgLazyFree"`
				PgLazyFreed         int64 `json:"pgLazyFreed"`
				PgMajFault          int64 `json:"pgMajFault"`
				PgRefill            int64 `json:"pgRefill"`
				PgScan              int64 `json:"pgScan"`
				PgSteal             int64 `json:"pgSteal"`
				Shmem               int64 `json:"shmem"`
				Slab                int64 `json:"slab"`
				SlabReclaimable     int64 `json:"slabReclaimable"`
				SlabUnreclaimable   int64 `json:"slabUnreclaimable"`
				Sock                int64 `json:"sock"`
				ThpCollapseAlloc    int64 `json:"thpCollapseAlloc"`
				ThpFaultAlloc       int64 `json:"thpFaultAlloc"`
				Unevictable         int64 `json:"unevictable"`
				WorkingSetActivate  int64 `json:"workingsetActivate"`
				WorkingSetNoReclaim int64 `json:"workingsetNodereclaim"`
				WorkingSetRefault   int64 `json:"workingsetRefault"`
			} `json:"stats"`
			Limit int64 `json:"limit"`
		}{
			Usage: 104857600,  // 100MB
			Limit: 1073741824, // 1GB
		},
		Networks: struct {
			Eth0 struct {
				RxBytes   int64 `json:"rxBytes"`
				RxPackets int64 `json:"rxPackets"`
				RxErrors  int64 `json:"rxErrors"`
				RxDropped int64 `json:"rxDropped"`
				TxBytes   int64 `json:"txBytes"`
				TxPackets int64 `json:"txPackets"`
				TxErrors  int64 `json:"txErrors"`
				TxDropped int64 `json:"txDropped"`
			} `json:"eth0"`
		}{
			Eth0: struct {
				RxBytes   int64 `json:"rxBytes"`
				RxPackets int64 `json:"rxPackets"`
				RxErrors  int64 `json:"rxErrors"`
				RxDropped int64 `json:"rxDropped"`
				TxBytes   int64 `json:"txBytes"`
				TxPackets int64 `json:"txPackets"`
				TxErrors  int64 `json:"txErrors"`
				TxDropped int64 `json:"txDropped"`
			}{
				RxBytes: 1048576, // 1MB
				TxBytes: 2097152, // 2MB
			},
		},
		BlkioStats: struct {
			IoServiceBytesRecursive []any `json:"ioServiceBytesRecursive"`
			IoServicedRecursive     []any `json:"ioServicedRecursive"`
			IoQueueRecursive        []any `json:"ioQueueRecursive"`
			IoServiceTimeRecursive  []any `json:"ioServiceTimeRecursive"`
			IoWaitTimeRecursive     []any `json:"ioWaitTimeRecursive"`
			IoMergedRecursive       []any `json:"ioMergedRecursive"`
			IoTimeRecursive         []any `json:"ioTimeRecursive"`
			SectorsRecursive        []any `json:"sectorsRecursive"`
		}{
			IoServiceBytesRecursive: []any{
				float64(1048576), // 1MB read
				float64(2097152), // 2MB write
			},
		},
	}

	// Convert stats to JSON
	statsJson, err := json.Marshal(stats)
	assert.NoError(t, err)

	// Write stats to formatter
	n, err := formatter.Write(statsJson)
	assert.NoError(t, err)
	assert.Equal(t, len(statsJson), n)

	// Parse formatted output
	var formatted FormatedContainerStats
	err = json.Unmarshal(buf.Bytes(), &formatted)
	assert.NoError(t, err)

	// Verify formatted values
	assert.Equal(t, "test-container", formatted.ID)
	assert.Equal(t, "test", formatted.Name)
	assert.Equal(t, "40.00%", formatted.CpuUsage) // (100000000/1000000000) * 4 * 100
	assert.Equal(t, "100.00 MB / 1.00 GB", formatted.MemoryUsage)
	assert.Equal(t, "1.00 MB / 2.00 MB", formatted.NetworkIO)
}

func TestFormatCpuUsagePercentage(t *testing.T) {
	tests := []struct {
		name     string
		stats    ContainerStats
		expected string
	}{
		{
			name: "Normal CPU Usage",
			stats: ContainerStats{
				CpuStats: struct {
					CpuUsage struct {
						TotalUsage        int64 `json:"totalUsage"`
						UsageInKernelMode int64 `json:"usageInKernelmode"`
						UsageInUserMode   int64 `json:"usageInUsermode"`
					} `json:"cpuUsage"`
					SystemCPUUsage int64 `json:"systemCpuUsage"`
					OnlineCPUs     int64 `json:"onlineCpus"`
					ThrottlingData struct {
						Periods          int64 `json:"periods"`
						ThrottledPeriods int64 `json:"throttledPeriods"`
						ThrottledTime    int64 `json:"throttledTime"`
					} `json:"throttlingData"`
				}{
					CpuUsage: struct {
						TotalUsage        int64 `json:"totalUsage"`
						UsageInKernelMode int64 `json:"usageInKernelmode"`
						UsageInUserMode   int64 `json:"usageInUsermode"`
					}{
						TotalUsage: 100000000,
					},
					SystemCPUUsage: 1000000000,
					OnlineCPUs:     4,
				},
				PreCPUStats: struct {
					CpuUsage struct {
						TotalUsage        int64 `json:"totalUsage"`
						UsageInKernelMode int64 `json:"usageInKernelmode"`
						UsageInUserMode   int64 `json:"usageInUsermode"`
					} `json:"cpuUsage"`
					SystemCPUUsage int64 `json:"systemCpuUsage"`
					OnlineCPUs     int64 `json:"onlineCpus"`
					ThrottlingData struct {
						Periods          int64 `json:"periods"`
						ThrottledPeriods int64 `json:"throttledPeriods"`
						ThrottledTime    int64 `json:"throttledTime"`
					} `json:"throttlingData"`
				}{
					CpuUsage: struct {
						TotalUsage        int64 `json:"totalUsage"`
						UsageInKernelMode int64 `json:"usageInKernelmode"`
						UsageInUserMode   int64 `json:"usageInUsermode"`
					}{
						TotalUsage: 0,
					},
					SystemCPUUsage: 0,
				},
			},
			expected: "40.00%",
		},
		{
			name: "Zero CPU Usage",
			stats: ContainerStats{
				CpuStats: struct {
					CpuUsage struct {
						TotalUsage        int64 `json:"totalUsage"`
						UsageInKernelMode int64 `json:"usageInKernelmode"`
						UsageInUserMode   int64 `json:"usageInUsermode"`
					} `json:"cpuUsage"`
					SystemCPUUsage int64 `json:"systemCpuUsage"`
					OnlineCPUs     int64 `json:"onlineCpus"`
					ThrottlingData struct {
						Periods          int64 `json:"periods"`
						ThrottledPeriods int64 `json:"throttledPeriods"`
						ThrottledTime    int64 `json:"throttledTime"`
					} `json:"throttlingData"`
				}{
					CpuUsage: struct {
						TotalUsage        int64 `json:"totalUsage"`
						UsageInKernelMode int64 `json:"usageInKernelmode"`
						UsageInUserMode   int64 `json:"usageInUsermode"`
					}{
						TotalUsage: 0,
					},
					SystemCPUUsage: 1000000000,
					OnlineCPUs:     4,
				},
				PreCPUStats: struct {
					CpuUsage struct {
						TotalUsage        int64 `json:"totalUsage"`
						UsageInKernelMode int64 `json:"usageInKernelmode"`
						UsageInUserMode   int64 `json:"usageInUsermode"`
					} `json:"cpuUsage"`
					SystemCPUUsage int64 `json:"systemCpuUsage"`
					OnlineCPUs     int64 `json:"onlineCpus"`
					ThrottlingData struct {
						Periods          int64 `json:"periods"`
						ThrottledPeriods int64 `json:"throttledPeriods"`
						ThrottledTime    int64 `json:"throttledTime"`
					} `json:"throttlingData"`
				}{
					CpuUsage: struct {
						TotalUsage        int64 `json:"totalUsage"`
						UsageInKernelMode int64 `json:"usageInKernelmode"`
						UsageInUserMode   int64 `json:"usageInUsermode"`
					}{
						TotalUsage: 0,
					},
					SystemCPUUsage: 0,
				},
			},
			expected: "0.00%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.stats.FormatCpuUsagePercentage()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBytesToHumanReadable(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"Bytes", 500, "500 B"},
		{"Kilobytes", 1500, "1.46 KB"},
		{"Megabytes", 1500000, "1.43 MB"},
		{"Gigabytes", 1500000000, "1.40 GB"},
		{"Terabytes", 1500000000000, "1.36 TB"},
		{"Zero", 0, "0 B"},
		{"Negative", -1000, "-1000 B"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bytesToHumanReadable(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatMemoryUsage(t *testing.T) {
	stats := ContainerStats{
		MemoryStats: struct {
			Usage int64 `json:"usage"`
			Stats struct {
				ActiveAnon          int64 `json:"activeAnon"`
				ActiveFile          int64 `json:"activeFile"`
				Anon                int64 `json:"anon"`
				AnonTHP             int64 `json:"anonTHP"`
				File                int64 `json:"file"`
				FileDirty           int64 `json:"fileDirty"`
				FileMapped          int64 `json:"fileMapped"`
				FileWriteBack       int64 `json:"fileWriteback"`
				InactiveAnon        int64 `json:"inactiveAnon"`
				InactiveFile        int64 `json:"inactiveFile"`
				KernelStack         int64 `json:"kernelStack"`
				PgActivate          int64 `json:"pgActivate"`
				PgDeactivate        int64 `json:"pgDeactivate"`
				PgFault             int64 `json:"pgFault"`
				PgLazyFree          int64 `json:"pgLazyFree"`
				PgLazyFreed         int64 `json:"pgLazyFreed"`
				PgMajFault          int64 `json:"pgMajFault"`
				PgRefill            int64 `json:"pgRefill"`
				PgScan              int64 `json:"pgScan"`
				PgSteal             int64 `json:"pgSteal"`
				Shmem               int64 `json:"shmem"`
				Slab                int64 `json:"slab"`
				SlabReclaimable     int64 `json:"slabReclaimable"`
				SlabUnreclaimable   int64 `json:"slabUnreclaimable"`
				Sock                int64 `json:"sock"`
				ThpCollapseAlloc    int64 `json:"thpCollapseAlloc"`
				ThpFaultAlloc       int64 `json:"thpFaultAlloc"`
				Unevictable         int64 `json:"unevictable"`
				WorkingSetActivate  int64 `json:"workingsetActivate"`
				WorkingSetNoReclaim int64 `json:"workingsetNodereclaim"`
				WorkingSetRefault   int64 `json:"workingsetRefault"`
			} `json:"stats"`
			Limit int64 `json:"limit"`
		}{
			Usage: 104857600,  // 100MB
			Limit: 1073741824, // 1GB
		},
	}

	result := stats.FormatMemoryUsage()
	assert.Equal(t, "100.00 MB / 1.00 GB", result)
}

func TestFormatNetworkIO(t *testing.T) {
	stats := ContainerStats{
		Networks: struct {
			Eth0 struct {
				RxBytes   int64 `json:"rxBytes"`
				RxPackets int64 `json:"rxPackets"`
				RxErrors  int64 `json:"rxErrors"`
				RxDropped int64 `json:"rxDropped"`
				TxBytes   int64 `json:"txBytes"`
				TxPackets int64 `json:"txPackets"`
				TxErrors  int64 `json:"txErrors"`
				TxDropped int64 `json:"txDropped"`
			} `json:"eth0"`
		}{
			Eth0: struct {
				RxBytes   int64 `json:"rxBytes"`
				RxPackets int64 `json:"rxPackets"`
				RxErrors  int64 `json:"rxErrors"`
				RxDropped int64 `json:"rxDropped"`
				TxBytes   int64 `json:"txBytes"`
				TxPackets int64 `json:"txPackets"`
				TxErrors  int64 `json:"txErrors"`
				TxDropped int64 `json:"txDropped"`
			}{
				RxBytes: 1048576, // 1MB
				TxBytes: 2097152, // 2MB
			},
		},
	}

	result := stats.FormatNetworkIO()
	assert.Equal(t, "1.00 MB / 2.00 MB", result)
}

func TestFormatDiskIO(t *testing.T) {
	stats := ContainerStats{
		BlkioStats: struct {
			IoServiceBytesRecursive []any `json:"ioServiceBytesRecursive"`
			IoServicedRecursive     []any `json:"ioServicedRecursive"`
			IoQueueRecursive        []any `json:"ioQueueRecursive"`
			IoServiceTimeRecursive  []any `json:"ioServiceTimeRecursive"`
			IoWaitTimeRecursive     []any `json:"ioWaitTimeRecursive"`
			IoMergedRecursive       []any `json:"ioMergedRecursive"`
			IoTimeRecursive         []any `json:"ioTimeRecursive"`
			SectorsRecursive        []any `json:"sectorsRecursive"`
		}{
			IoServiceBytesRecursive: []any{
				float64(1048576), // 1MB read
				float64(2097152), // 2MB write
			},
		},
	}

	result := stats.FormatDiskIO()
	assert.Equal(t, "1.00 MB / 2.00 MB", result)
}
