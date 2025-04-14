package config

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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

func (c Config) PackagesToTest() ([]string, error) {
	cmd := exec.Command("go", append([]string{"list"}, c.Include...)...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	all := strings.Split(strings.TrimSpace(string(out)), "\n")
	var pkgs []string

	for _, pkg := range all {
		skip := false
		for _, exclude := range c.Exclude {
			if strings.HasPrefix(pkg, exclude) {
				skip = true
				break
			}
		}
		if !skip {
			pkgs = append(pkgs, pkg)
		}
	}

	return pkgs, nil
}
