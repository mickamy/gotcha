package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const Path = ".gotcha.yaml"

type Config struct {
	Include   []string `yaml:"include"`
	Exclude   []string `yaml:"exclude"`
	TestFlags []string `yaml:"test_flags"`
}

func Default() *Config {
	return &Config{
		Include:   []string{"./..."},
		Exclude:   []string{"vendor/", "mocks/"},
		TestFlags: []string{"-v"},
	}
}

func Load() (Config, error) {
	return LoadByPath(Path)
}

func LoadByPath(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := new(Config)
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}

	return *cfg, nil
}
