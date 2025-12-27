package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/toms74209200/gh-atat/internal/config"
)

// ConfigStorage is an abstract configuration persistence interface
type ConfigStorage interface {
	// LoadConfig loads configuration into a map.
	// This method should handle parsing of the configuration file content.
	LoadConfig() (map[config.ConfigKey]any, error)

	// SaveConfig saves the given configuration map.
	// This method should handle serializing the map to the appropriate format before writing.
	SaveConfig(configData map[config.ConfigKey]any) error
}

// LocalConfigStorage is a file-based local configuration persistence implementation
type LocalConfigStorage struct {
	configPath string
	configDir  string
}

// NewLocalConfigStorage creates a new LocalConfigStorage instance
func NewLocalConfigStorage() (*LocalConfigStorage, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	configDir := filepath.Join(currentDir, config.ProjectConfigDir)
	configPath := filepath.Join(configDir, config.ProjectConfigFilename)

	return &LocalConfigStorage{
		configPath: configPath,
		configDir:  configDir,
	}, nil
}

// LoadConfig loads configuration from the local config file
func (s *LocalConfigStorage) LoadConfig() (map[config.ConfigKey]any, error) {
	content, err := readFileBytes(s.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project config file at %s: %w", s.configPath, err)
	}

	return config.ParseConfig(content)
}

// SaveConfig saves configuration to the local config file
func (s *LocalConfigStorage) SaveConfig(configData map[config.ConfigKey]any) error {
	// Create config directory if it doesn't exist
	if _, err := os.Stat(s.configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(s.configDir, 0755); err != nil {
			return fmt.Errorf("failed to create project config directory at %s: %w", s.configDir, err)
		}
	}

	// Convert map to JSON
	jsonMap := make(map[string]any)
	for key, value := range configData {
		jsonMap[string(key)] = value
	}

	// Serialize to pretty JSON
	contentBytes, err := json.MarshalIndent(jsonMap, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config to JSON for saving: %w", err)
	}

	// Write to file
	if err := os.WriteFile(s.configPath, contentBytes, 0644); err != nil {
		return fmt.Errorf("failed to write to project config file at %s: %w", s.configPath, err)
	}

	return nil
}

// readFileBytes reads the content of the file at the specified path into a byte slice.
//
// - Returns an empty byte slice if the file does not exist.
// - Returns an error if any other error occurs during reading (e.g., permission denied).
func readFileBytes(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []byte{}, nil
		}
		return nil, fmt.Errorf("failed to read file: %s: %w", path, err)
	}
	return content, nil
}
