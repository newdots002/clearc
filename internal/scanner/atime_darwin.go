//go:build darwin

package scanner

import (
	"os"
	"syscall"
	"time"
)

func getAccessTime(info os.FileInfo) time.Time {
	stat := info.Sys().(*syscall.Stat_t)
	return time.Unix(stat.Atimespec.Sec, stat.Atimespec.Nsec)
}
