package config

import (
	"fmt"
	"os"

	"github.com/muasiq/universal-inverter-data-normalizer/internal/provider"
	"gopkg.in/yaml.v3"
)

// Config is the top-level application configuration.
type Config struct {
	Server    ServerConfig            `yaml:"server"`
	Providers []provider.ProviderConfig `yaml:"providers"`
	Logging   LoggingConfig           `yaml:"logging"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type LoggingConfig struct {
	Level  string `yaml:"level"`  // "debug", "info", "warn", "error"
	Format string `yaml:"format"` // "json" or "console"
}

// Load reads and parses a YAML configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %s: %w", path, err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "console",
		},
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file %s: %w", path, err)
	}

	return cfg, nil
}
