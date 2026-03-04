package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ServerConfig holds HTTP server settings for web mode
type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

// Config holds all application configuration from config.json
type Config struct {
	Server ServerConfig `json:"server"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8090,
			Host: "127.0.0.1",
		},
	}
}

// configPath returns the path to config.json
func configPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".dbbackup", "config.json"), nil
}

// Load reads config from ~/.dbbackup/config.json, returns defaults if not found
func Load() (*Config, error) {
	cfg := DefaultConfig()

	path, err := configPath()
	if err != nil {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		fmt.Printf("Warning: could not read config file %s: %v\n", path, err)
		return cfg, nil
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return DefaultConfig(), nil
	}

	// Ensure sensible defaults for missing values
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8090
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "127.0.0.1"
	}

	return cfg, nil
}

// Save writes config to ~/.dbbackup/config.json
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
