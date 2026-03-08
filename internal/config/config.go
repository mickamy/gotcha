package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const DefaultPath = ".gotcha.yaml"

// Config holds gotcha settings loaded from YAML.
type Config struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
	Args    []string `yaml:"args"`
}

// Default returns sensible defaults.
func Default() Config {
	return Config{
		Include: []string{"./..."},
		Exclude: []string{"vendor/", "mocks/"},
		Args:    []string{"-v"},
	}
}

// Load reads the config from the default path.
func Load() (Config, error) {
	return LoadByPath(DefaultPath)
}

// LoadByPath reads the config from the given file.
func LoadByPath(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

// Save writes the config as YAML to the given path.
func Save(path string, cfg Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// ShouldExclude reports whether the given path matches any exclusion pattern.
func (c Config) ShouldExclude(path string) bool {
	for _, ex := range c.Exclude {
		if strings.Contains(path, ex) {
			return true
		}
	}
	return false
}
