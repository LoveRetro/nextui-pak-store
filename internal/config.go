package internal

import (
	"encoding/json"
	"os"
	"path/filepath"

	gaba "github.com/BrandonKowalski/gabagool/v2/pkg/gabagool"
	"github.com/LoveRetro/nextui-pak-store/utils"
)

type PlatformFilterMode string

const (
	PlatformFilterMatchDevice PlatformFilterMode = "match_device"
	PlatformFilterAll         PlatformFilterMode = "all"
)

type DebugLevel string

const (
	DebugLevelError DebugLevel = "error"
	DebugLevelInfo  DebugLevel = "info"
	DebugLevelDebug DebugLevel = "debug"
)

type Config struct {
	PlatformFilter           PlatformFilterMode `json:"platform_filter"`
	DebugLevel               DebugLevel         `json:"debug_level"`
	DiscoverExistingInstalls *bool              `json:"discover_existing_installs"`
}

var configInstance *Config

func LoadConfig() (*Config, error) {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			config := &Config{
				PlatformFilter: PlatformFilterMatchDevice,
				DebugLevel:     DebugLevelError,
			}
			configInstance = config
			return config, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.PlatformFilter == "" {
		config.PlatformFilter = PlatformFilterMatchDevice
	}
	if config.DebugLevel == "" {
		config.DebugLevel = DebugLevelError
	}

	configInstance = &config
	return &config, nil
}

func SaveConfig(config *Config) error {
	logger := gaba.GetLogger()

	if config.PlatformFilter == "" {
		config.PlatformFilter = PlatformFilterMatchDevice
	}
	if config.DebugLevel == "" {
		config.DebugLevel = DebugLevelError
	}

	configPath := getConfigPath()

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		logger.Error("Failed to create config directory", "error", err)
		return err
	}

	pretty, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal config to JSON", "error", err)
		return err
	}

	if err := os.WriteFile(configPath, pretty, 0644); err != nil {
		logger.Error("Failed to write config file", "error", err)
		return err
	}

	configInstance = config
	return nil
}

func GetConfig() *Config {
	if configInstance == nil {
		config, err := LoadConfig()
		if err != nil {
			return &Config{PlatformFilter: PlatformFilterMatchDevice, DebugLevel: DebugLevelError}
		}
		return config
	}
	return configInstance
}

func getConfigPath() string {
	if os.Getenv("ENVIRONMENT") == "DEV" {
		return "config.json"
	}
	return filepath.Join(utils.GetUserDataDir(), "config.json")
}

func (c *Config) ShouldDiscoverExistingInstalls() bool {
	if c.DiscoverExistingInstalls == nil {
		return true
	}
	return *c.DiscoverExistingInstalls
}
