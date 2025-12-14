package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ServerType represents the type of MCP server
type ServerType string

const (
	ServerTypeNPM  ServerType = "npm"
	ServerTypePIP  ServerType = "pip"
	ServerTypeCustom ServerType = "custom"
)

// ServerStatus represents the status of an MCP server
type ServerStatus string

const (
	StatusInstalled ServerStatus = "installed"
	StatusRunning   ServerStatus = "running"
	StatusStopped   ServerStatus = "stopped"
	StatusError     ServerStatus = "error"
)

// ServerConfig represents the configuration for an installed MCP server
type ServerConfig struct {
	Name         string                 `json:"name"`
	Type         ServerType            `json:"type"`
	Package      string                 `json:"package,omitempty"`
	Version      string                 `json:"version,omitempty"`
	InstalledAt  time.Time             `json:"installed_at"`
	Status       ServerStatus          `json:"status"`
	Config       map[string]interface{} `json:"config,omitempty"`
	Dependencies map[string]string     `json:"dependencies,omitempty"`
	Path         string                 `json:"path,omitempty"` // Installation path
}

// Credential represents API keys and credentials for a server
type Credential struct {
	Data map[string]string `json:"data"`
}

// FileStorage provides filesystem-based storage for MCP server data
type FileStorage struct {
	baseDir string
}

// NewFileStorage creates a new filesystem-based storage instance
func NewFileStorage(baseDir string) (*FileStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{"servers", "credentials", "logs", "cache"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(baseDir, dir), 0755); err != nil {
			return nil, fmt.Errorf("failed to create %s directory: %w", dir, err)
		}
	}

	return &FileStorage{baseDir: baseDir}, nil
}

// SaveServerConfig saves a server configuration
func (fs *FileStorage) SaveServerConfig(server *ServerConfig) error {
	data, err := json.MarshalIndent(server, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal server config: %w", err)
	}

	filename := fmt.Sprintf("%s.json", server.Name)
	path := filepath.Join(fs.baseDir, "servers", filename)

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write server config: %w", err)
	}

	return nil
}

// LoadServerConfig loads a server configuration
func (fs *FileStorage) LoadServerConfig(name string) (*ServerConfig, error) {
	filename := fmt.Sprintf("%s.json", name)
	path := filepath.Join(fs.baseDir, "servers", filename)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("server %s not found", name)
		}
		return nil, fmt.Errorf("failed to read server config: %w", err)
	}

	var config ServerConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse server config: %w", err)
	}

	return &config, nil
}

// ListServerConfigs returns all server configurations
func (fs *FileStorage) ListServerConfigs() ([]*ServerConfig, error) {
	fmt.Printf("DEBUG: ListServerConfigs called, baseDir: %s\n", fs.baseDir)
	pattern := filepath.Join(fs.baseDir, "servers", "*.json")
	fmt.Printf("DEBUG: Pattern: %s\n", pattern)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list server configs: %w", err)
	}

	fmt.Printf("DEBUG: Found %d matches\n", len(matches))
	var configs []*ServerConfig
	for _, match := range matches {
		fmt.Printf("DEBUG: Reading %s\n", match)
		data, err := os.ReadFile(match)
		if err != nil {
			fmt.Printf("DEBUG: Error reading %s: %v\n", match, err)
			continue // Skip files that can't be read
		}

		var config ServerConfig
		if err := json.Unmarshal(data, &config); err != nil {
			fmt.Printf("DEBUG: Error unmarshaling %s: %v\n", match, err)
			continue // Skip invalid files
		}

		configs = append(configs, &config)
	}

	fmt.Printf("DEBUG: Returning %d configs\n", len(configs))
	return configs, nil
}

// DeleteServerConfig deletes a server configuration
func (fs *FileStorage) DeleteServerConfig(name string) error {
	filename := fmt.Sprintf("%s.json", name)
	path := filepath.Join(fs.baseDir, "servers", filename)

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete server config: %w", err)
	}

	return nil
}

// SaveCredentials saves credentials for a server
func (fs *FileStorage) SaveCredentials(name string, creds *Credential) error {
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	filename := fmt.Sprintf("%s.key", name)
	path := filepath.Join(fs.baseDir, "credentials", filename)

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

// LoadCredentials loads credentials for a server
func (fs *FileStorage) LoadCredentials(name string) (*Credential, error) {
	filename := fmt.Sprintf("%s.key", name)
	path := filepath.Join(fs.baseDir, "credentials", filename)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("credentials for %s not found", name)
		}
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var creds Credential
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &creds, nil
}

// DeleteCredentials deletes credentials for a server
func (fs *FileStorage) DeleteCredentials(name string) error {
	filename := fmt.Sprintf("%s.key", name)
	path := filepath.Join(fs.baseDir, "credentials", filename)

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}

	return nil
}

// GetLogPath returns the log file path for a server
func (fs *FileStorage) GetLogPath(name string) string {
	return filepath.Join(fs.baseDir, "logs", fmt.Sprintf("%s.log", name))
}

// GetCacheDir returns the cache directory
func (fs *FileStorage) GetCacheDir() string {
	return filepath.Join(fs.baseDir, "cache")
}

// GetServersDir returns the servers directory
func (fs *FileStorage) GetServersDir() string {
	return filepath.Join(fs.baseDir, "servers")
}

// GetCredentialsDir returns the credentials directory
func (fs *FileStorage) GetCredentialsDir() string {
	return filepath.Join(fs.baseDir, "credentials")
}

// GetLogsDir returns the logs directory
func (fs *FileStorage) GetLogsDir() string {
	return filepath.Join(fs.baseDir, "logs")
}