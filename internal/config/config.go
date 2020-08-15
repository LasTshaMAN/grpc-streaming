package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Config contains all configuration settings.
type Config struct {
	URLs             []string      `yaml:"URLs"`
	MinTimeout       time.Duration `yaml:"MinTimeout"`
	MaxTimeout       time.Duration `yaml:"MaxTimeout"`
	NumberOfRequests int           `yaml:"NumberOfRequests"`
}

// Parse YAML configuration file.
func Parse(filePath string) (Config, error) {
	c := Config{}

	f, err := os.Open(filePath)
	if err != nil {
		return c, fmt.Errorf("cannot read from file: %s, err: %w", filePath, err)
	}

	d := yaml.NewDecoder(f)
	d.SetStrict(true)

	err = d.Decode(&c)
	if err != nil {
		return c, fmt.Errorf("error decoding file: %s, err: %w", filePath, err)
	}

	err = f.Close()
	if err != nil {
		return c, fmt.Errorf("cannot close file: %s, err: %w", filePath, err)
	}

	return c, nil
}
