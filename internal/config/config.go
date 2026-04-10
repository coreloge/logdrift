package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ServiceConfig defines a single log source to tail.
type ServiceConfig struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	File    string `yaml:"file"`
	Color   string `yaml:"color"`
}

// Config is the top-level logdrift configuration.
type Config struct {
	Services []ServiceConfig `yaml:"services"`
	DiffMode bool            `yaml:"diff_mode"`
	MaxLines int             `yaml:"max_lines"`
}

// DefaultMaxLines is used when max_lines is not specified.
const DefaultMaxLines = 500

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	if cfg.MaxLines == 0 {
		cfg.MaxLines = DefaultMaxLines
	}

	return &cfg, nil
}

// validate checks that the config is semantically valid.
func (c *Config) validate() error {
	if len(c.Services) == 0 {
		return fmt.Errorf("config must define at least one service")
	}
	names := make(map[string]struct{}, len(c.Services))
	for i, svc := range c.Services {
		if svc.Name == "" {
			return fmt.Errorf("service[%d]: name is required", i)
		}
		if svc.Command == "" && svc.File == "" {
			return fmt.Errorf("service %q: either command or file must be set", svc.Name)
		}
		if _, dup := names[svc.Name]; dup {
			return fmt.Errorf("duplicate service name %q", svc.Name)
		}
		names[svc.Name] = struct{}{}
	}
	return nil
}
