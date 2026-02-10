//go:build !windows

package scanner

import (
	"os"
	"syscall"
	"time"
)

func getAccessTime(info os.FileInfo) time.Time {
	stat := info.Sys().(*syscall.Stat_t)
	return time.Unix(stat.Atim.Sec, stat.Atim.Nsec)
}
