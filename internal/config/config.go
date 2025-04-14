package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mickamy/gotcha/internal/paths"
)

const Path = ".gotcha.yaml"

type Config struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
	Args    []string `yaml:"args"`
}

func Default() *Config {
	return &Config{
		Include: []string{"./..."},
		Exclude: []string{"vendor/", "mocks/"},
		Args:    []string{"-v"},
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
	all, err := paths.ListPackages(c.Include)
	if err != nil {
		return nil, err
	}

	modulePath, err := paths.ModulePath()
	if err != nil {
		return nil, err
	}

	var pkgs []string
	for _, pkg := range all {
		if c.ShouldExclude(pkg) {
			continue
		}
		if strings.HasPrefix(pkg, modulePath) {
			rel := strings.TrimPrefix(pkg, modulePath)
			if rel == "" {
				pkgs = append(pkgs, ".")
			} else {
				pkgs = append(pkgs, "./"+strings.TrimPrefix(rel, "/"))
			}
		} else {
			pkgs = append(pkgs, pkg)
		}
	}
	return pkgs, nil
}

func (c Config) ShouldExclude(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	for _, ex := range c.Exclude {
		absEx, err := filepath.Abs(ex)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absPath, absEx) {
			return true
		}
	}
	return false
}
