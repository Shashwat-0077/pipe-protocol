package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config holds the server configuration
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
	Logging LoggingConfig `yaml:"logging"`
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port           int    `yaml:"port"`
	Host           string `yaml:"host"`
	ServerName     string `yaml:"server_name"`
	MaxConnections int    `yaml:"max_connections"`
	Timeout        int    `yaml:"timeout_seconds"`
}

// StorageConfig holds storage configuration
type StorageConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// LoadConfig loads configuration from file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// Set defaults
	if config.Server.Port == 0 {
		config.Server.Port = 2525
	}
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}
	if config.Server.MaxConnections == 0 {
		config.Server.MaxConnections = 100
	}
	if config.Server.Timeout == 0 {
		config.Server.Timeout = 30
	}
	if config.Storage.Type == "" {
		config.Storage.Type = "file"
	}
	if config.Storage.Path == "" {
		config.Storage.Path = "./data"
	}

	return &config, nil
}
