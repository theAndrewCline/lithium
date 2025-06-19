package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	DatabasePath string  `yaml:"databasePath"`
	SyncUrl      *string `yaml:"syncUrl"`
}

// DefaultConfig returns a config with default values
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("could not get user home dir")
	}

	dbPath := fmt.Sprintf("file:%s", filepath.Join(homeDir, ".lithium", "tasks.db"))

	fmt.Println(dbPath)

	return &Config{DatabasePath: dbPath}
}

// LoadConfig loads configuration from the standard config locations
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	configPath, err := findConfigFile()
	if err != nil {
		// Config file not found, use defaults
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// findConfigFile looks for config file in standard locations
func findConfigFile() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Possible config file locations
	configPaths := []string{
		filepath.Join(homeDir, ".config", "lithium", "config.yaml"),
		filepath.Join(homeDir, ".config", "lithium", "config.yml"),
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("config file not found")
}

// EnsureDatabaseDir creates the database directory if it doesn't exist
func (c *Config) EnsureDatabaseDir() error {
	dbDir := filepath.Dir(c.DatabasePath)
	return os.MkdirAll(dbDir, 0755)
}
