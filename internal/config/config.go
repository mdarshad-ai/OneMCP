package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultMCPDir = ".mcp"
	ConfigFile    = "config.json"
)

// Config represents the global configuration
type Config struct {
	Version   string     `json:"version"`
	Gateway   GatewayConfig `json:"gateway"`
	Web       WebConfig   `json:"web"`
	AutoUpdate bool       `json:"auto_update"`
	LogLevel   string     `json:"log_level"`
}

// GatewayConfig holds gateway-specific settings
type GatewayConfig struct {
	Port   int    `json:"port"`
	Host   string `json:"host"`
	Transport string `json:"transport"`
}

// WebConfig holds web interface settings
type WebConfig struct {
	Port   int    `json:"port"`
	Host   string `json:"host"`
	Enabled bool  `json:"enabled"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Version: "0.1.0",
		Gateway: GatewayConfig{
			Port:     5234,
			Host:     "127.0.0.1",
			Transport: "stdio",
		},
		Web: WebConfig{
			Port:    80,
			Host:    "0.0.0.0",
			Enabled: true,
		},
		AutoUpdate: true,
		LogLevel:   "info",
	}
}

// LoadConfig loads configuration from the MCP directory
func LoadConfig(mcpDir string) (*Config, error) {
	configPath := filepath.Join(mcpDir, ConfigFile)

	// Create default config if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig()
		if err := SaveConfig(mcpDir, config); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return config, nil
	}

	// Load existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to the MCP directory
func SaveConfig(mcpDir string, config *Config) error {
	if err := os.MkdirAll(mcpDir, 0755); err != nil {
		return fmt.Errorf("failed to create MCP directory: %w", err)
	}

	configPath := filepath.Join(mcpDir, ConfigFile)
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetMCPDir returns the MCP directory path
func GetMCPDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, DefaultMCPDir), nil
}