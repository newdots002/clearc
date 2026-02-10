//go:build linux

package platform

import (
	"os"
	"path/filepath"
)

// GetSystemDrive returns the system drive on Linux
func GetSystemDrive() string {
	return "/"
}

// GetUserHome returns the user's home directory
func GetUserHome() string {
	return os.Getenv("HOME")
}

// GetTempDirs returns temporary directories on Linux
func GetTempDirs() []string {
	userHome := GetUserHome()

	return []string{
		"/tmp",
		"/var/tmp",
		filepath.Join(userHome, ".cache"),
	}
}

// GetCacheDirs returns cache directories on Linux
func GetCacheDirs() []string {
	userHome := GetUserHome()

	return []string{
		filepath.Join(userHome, ".cache"),
		filepath.Join(userHome, ".local", "share", "Trash"),
		filepath.Join(userHome, ".config", "google-chrome", "Default", "Cache"),
		filepath.Join(userHome, ".mozilla", "firefox"),
	}
}

// GetTrashDir returns the trash directory on Linux
func GetTrashDir() string {
	return filepath.Join(GetUserHome(), ".local", "share", "Trash")
}

// GetUserDataDir returns the user data directory
func GetUserDataDir() string {
	return filepath.Join(GetUserHome(), ".config", "clearc")
}

// GetDevCacheDirs returns development-related cache directories
func GetDevCacheDirs() map[string]DevCacheInfo {
	userHome := GetUserHome()

	return map[string]DevCacheInfo{
		"node_modules": {
			Name:        "Node.js 依赖缓存",
			Description: "node_modules 目录",
			Patterns:    []string{"node_modules"},
			SearchPaths: []string{userHome},
			Color:       "#3B82F6",
		},
		"npm_cache": {
			Name:        "NPM 缓存",
			Description: "NPM 包缓存目录",
			Patterns:    []string{".npm"},
			SearchPaths: []string{userHome},
			Color:       "#3B82F6",
		},
		"go_cache": {
			Name:        "Go 模块缓存",
			Description: "go/pkg/mod/cache 目录",
			Patterns:    []string{"go/pkg/mod/cache"},
			SearchPaths: []string{userHome},
			Color:       "#F59E0B",
		},
		"python_cache": {
			Name:        "Python 缓存",
			Description: "__pycache__ 和 .pyc 文件",
			Patterns:    []string{"__pycache__", "*.pyc", ".pytest_cache", ".venv", "venv"},
			SearchPaths: []string{userHome},
			Color:       "#22C55E",
		},
		"rust_target": {
			Name:        "Rust 构建缓存",
			Description: "target 目录",
			Patterns:    []string{"target"},
			SearchPaths: []string{userHome},
			Color:       "#EF4444",
		},
		"docker_cache": {
			Name:        "Docker 缓存",
			Description: "Docker 构建缓存",
			Patterns:    []string{".docker"},
			SearchPaths: []string{userHome},
			Color:       "#3B82F6",
		},
	}
}

// DevCacheInfo represents information about a development cache category
type DevCacheInfo struct {
	Name        string
	Description string
	Patterns    []string
	SearchPaths []string
	Color       string
}
