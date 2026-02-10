package platform

import (
	"github.com/shirou/gopsutil/v3/disk"
)

// DiskUsageInfo represents disk usage information
type DiskUsageInfo struct {
	Total       uint64
	Used        uint64
	Free        uint64
	UsedPercent float64
}

// GetDiskUsage returns disk usage for a given path
func GetDiskUsage(path string) (*DiskUsageInfo, error) {
	usage, err := disk.Usage(path)
	if err != nil {
		return nil, err
	}
	return &DiskUsageInfo{
		Total:       usage.Total,
		Used:        usage.Used,
		Free:        usage.Free,
		UsedPercent: usage.UsedPercent,
	}, nil
}

// Platform interface defines platform-specific operations
type Platform interface {
	GetTempDirs() []string
	GetCacheDirs() []string
	GetTrashDir() string
	GetUserDataDir() string
	GetSystemDrive() string
	GetUserHome() string
}
