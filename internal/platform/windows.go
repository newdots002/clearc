//go:build windows

package platform

import (
	"os"
	"path/filepath"
)

// GetSystemDrive returns the system drive (e.g., "C:")
func GetSystemDrive() string {
	return os.Getenv("SystemDrive")
}

// GetUserHome returns the user's home directory
func GetUserHome() string {
	return os.Getenv("USERPROFILE")
}

// GetTempDirs returns temporary directories on Windows
func GetTempDirs() []string {
	userHome := GetUserHome()
	systemDrive := GetSystemDrive()

	return []string{
		os.Getenv("TEMP"),
		os.Getenv("TMP"),
		filepath.Join(systemDrive, "Windows", "Temp"),
		filepath.Join(userHome, "AppData", "Local", "Temp"),
	}
}

// GetCacheDirs returns cache directories on Windows
func GetCacheDirs() []string {
	userHome := GetUserHome()

	return []string{
		filepath.Join(userHome, "AppData", "Local", "Microsoft", "Windows", "INetCache"),
		filepath.Join(userHome, "AppData", "Local", "Microsoft", "Windows", "Explorer"),
		filepath.Join(userHome, "AppData", "Local", "Google", "Chrome", "User Data", "Default", "Cache"),
		filepath.Join(userHome, "AppData", "Local", "Microsoft", "Edge", "User Data", "Default", "Cache"),
		filepath.Join(userHome, "AppData", "Local", "Mozilla", "Firefox", "Profiles"),
	}
}

// GetTrashDir returns the recycle bin directory on Windows
func GetTrashDir() string {
	return filepath.Join(GetSystemDrive(), "$Recycle.Bin")
}

// GetUserDataDir returns the user data directory
func GetUserDataDir() string {
	return filepath.Join(GetUserHome(), "AppData", "Local", "ClearC")
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
		"yarn_cache": {
			Name:        "Yarn 缓存",
			Description: "Yarn 包缓存目录",
			Patterns:    []string{".yarn", ".yarnrc"},
			SearchPaths: []string{userHome, filepath.Join(userHome, "AppData", "Local", "Yarn")},
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
			Patterns:    []string{"__pycache__", "*.pyc", ".pytest_cache"},
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
		"maven_cache": {
			Name:        "Maven 缓存",
			Description: ".m2/repository 目录",
			Patterns:    []string{".m2/repository"},
			SearchPaths: []string{userHome},
			Color:       "#8B5CF6",
		},
		"gradle_cache": {
			Name:        "Gradle 缓存",
			Description: ".gradle 目录",
			Patterns:    []string{".gradle"},
			SearchPaths: []string{userHome},
			Color:       "#8B5CF6",
		},
		"ide_cache": {
			Name:        "IDE 缓存",
			Description: ".idea, .vscode 等目录",
			Patterns:    []string{".idea", ".vscode", "*.swp", "*.swo"},
			SearchPaths: []string{userHome},
			Color:       "#6B7280",
		},
		"build_output": {
			Name:        "构建产物",
			Description: "dist, build, out 目录",
			Patterns:    []string{"dist", "build", "out", ".next", ".nuxt"},
			SearchPaths: []string{userHome},
			Color:       "#6B7280",
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
