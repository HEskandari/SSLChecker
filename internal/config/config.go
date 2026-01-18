package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadConfig reads and parses the YAML configuration file
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate required fields
	if len(cfg.Domains) == 0 {
		return nil, fmt.Errorf("no domains configured")
	}
	for i, d := range cfg.Domains {
		if d.Host == "" {
			return nil, fmt.Errorf("domain %d missing host", i)
		}
		if d.Port == 0 {
			cfg.Domains[i].Port = 443
		}
	}

	// Ensure state file path is absolute
	if cfg.State.File != "" && !filepath.IsAbs(cfg.State.File) {
		absPath, err := filepath.Abs(cfg.State.File)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve state file path: %w", err)
		}
		cfg.State.File = absPath
	}

	return cfg, nil
}