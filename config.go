package main

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// Modem represents the address of the modem and its admin credentials.
type Modem struct {
	Address  string `yaml:"address"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Telemetry represents the exporter's listen address and metrics URI path.
type Telemetry struct {
	ListenAddress string `yaml:"listen_address"`
	MetricsPath   string `yaml:"metrics_path"`
}

// Config represents the yaml config file structure.
type Config struct {
	Modem     Modem     `yaml:"modem"`
	Telemetry Telemetry `yaml:"telemetry"`
}

// NewConfigFromFile reads the configuration file from the given path
// and returns a populated Config struct.
func NewConfigFromFile(path string) (*Config, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config file")
	}

	// Setup default config.
	config := Config{
		Modem: Modem{
			Address:  "192.168.100.1",
			Username: "admin",
		},
		Telemetry: Telemetry{
			ListenAddress: ":9527",
			MetricsPath:   "/metrics",
		},
	}

	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, errors.Wrap(err, "unable to parse config YAML")
	}

	if config.Modem.Password == "" {
		return nil, fmt.Errorf("modem password isn't set in config")
	}

	return &config, nil
}
