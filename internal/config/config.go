package config

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
