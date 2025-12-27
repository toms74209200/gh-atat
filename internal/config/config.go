package config

import (
	"encoding/json"
	"fmt"
	"maps"
)

// ConfigKey represents configuration keys
type ConfigKey string

const (
	// Repositories is the key for repository configuration
	Repositories ConfigKey = "repositories"
)

// Constants for configuration file paths
const (
	// ProjectConfigFilename is the filename for project-specific configuration
	ProjectConfigFilename = "config.json"
	// ProjectConfigDir is the directory name for project-specific configuration
	ProjectConfigDir = ".atat"
)

// AllConfigKeys returns all available configuration keys
func AllConfigKeys() []ConfigKey {
	return []ConfigKey{Repositories}
}

// ParseConfig parses a JSON configuration file content into a map of configuration values.
//
// Expects content to be a byte slice representing a JSON object with configuration keys.
// Returns a map of ConfigKey to any containing all parsed configuration values.
// Returns an empty map if the input content is empty or contains only whitespace.
// Returns an error if the JSON parsing fails.
func ParseConfig(content []byte) (map[ConfigKey]any, error) {
	// Check if content is empty or only whitespace
	if len(content) == 0 || isWhitespace(content) {
		return make(map[ConfigKey]any), nil
	}

	// Parse JSON content
	var value any
	if err := json.Unmarshal(content, &value); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	configMap := make(map[ConfigKey]any)

	// Check if value is an object
	if obj, ok := value.(map[string]any); ok {
		for _, key := range AllConfigKeys() {
			if val, exists := obj[string(key)]; exists {
				configMap[key] = val
			}
		}
		return configMap, nil
	}

	return nil, fmt.Errorf("config must be a JSON object")
}

// UpdateConfig merges updates into baseConfig and returns a new configuration map.
//
// Keys from updates are added to a clone of baseConfig.
// If a key exists in both, the value from updates overwrites the value in the cloned baseConfig.
//
// Returns a new map representing the merged configuration.
func UpdateConfig(baseConfig map[ConfigKey]any, updates map[ConfigKey]any) map[ConfigKey]any {
	newConfig := make(map[ConfigKey]any, len(baseConfig)+len(updates))

	// Copy base config
	maps.Copy(newConfig, baseConfig)

	// Apply updates
	maps.Copy(newConfig, updates)

	return newConfig
}

// isWhitespace checks if all bytes in the slice are ASCII whitespace
func isWhitespace(content []byte) bool {
	for _, b := range content {
		if b != ' ' && b != '\t' && b != '\n' && b != '\r' {
			return false
		}
	}
	return true
}
