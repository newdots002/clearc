package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"clearc/internal/platform"
)

// Config represents the application configuration
type Config struct {
	// General settings
	StartOnBoot       bool `json:"startOnBoot"`
	MinimizeToTray    bool `json:"minimizeToTray"`
	ShowNotifications bool `json:"showNotifications"`
	ScanReminder      bool `json:"scanReminder"`
	ReminderDays      int  `json:"reminderDays"`

	// Scan settings
	ScanNodeModules  bool `json:"scanNodeModules"`
	ScanGoCache      bool `json:"scanGoCache"`
	ScanPythonCache  bool `json:"scanPythonCache"`
	ScanRustTarget   bool `json:"scanRustTarget"`
	ScanTempFiles    bool `json:"scanTempFiles"`
	ScanBrowserCache bool `json:"scanBrowserCache"`
	ScanIDECache     bool `json:"scanIDECache"`
	ScanBuildOutput  bool `json:"scanBuildOutput"`

	// Unused files settings
	UnusedDaysThreshold int   `json:"unusedDaysThreshold"`
	UnusedMinSizeMB     int64 `json:"unusedMinSizeMB"`

	// UI settings
	Theme string `json:"theme"`
}

// New creates a new Config with default values
func New() *Config {
	return &Config{
		StartOnBoot:         true,
		MinimizeToTray:      true,
		ShowNotifications:   true,
		ScanReminder:        false,
		ReminderDays:        7,
		ScanNodeModules:     true,
		ScanGoCache:         true,
		ScanPythonCache:     true,
		ScanRustTarget:      true,
		ScanTempFiles:       true,
		ScanBrowserCache:    false,
		ScanIDECache:        false,
		ScanBuildOutput:     false,
		UnusedDaysThreshold: 30,
		UnusedMinSizeMB:     100,
		Theme:               "dark",
	}
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	dataDir := platform.GetUserDataDir()
	return filepath.Join(dataDir, "config.json")
}

// Load loads the configuration from disk
func (c *Config) Load() error {
	configPath := GetConfigPath()

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Save default config
		return c.Save()
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// Parse JSON
	return json.Unmarshal(data, c)
}

// Save saves the configuration to disk
func (c *Config) Save() error {
	configPath := GetConfigPath()

	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// Convert to JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(configPath, data, 0644)
}
