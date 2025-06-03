package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	MetricsAddress     string `yaml:"metricsAddress,omitempty"`
	HealthProbeAddress string `yaml:"healthProbeAddress,omitempty"`
	LeaderElection     bool   `yaml:"leaderElection,omitempty"`
	InstanceName       string `yaml:"instanceName,omitempty"`
	SilenceAuthor      string `yaml:"silenceAuthor,omitempty"`
	Interval           int    `yaml:"interval,omitempty"`
	Duration           int    `yaml:"duration,omitempty"`
	Concurrency        int    `yaml:"concurrency,omitempty"`
}

func LoadConfig(path string) (*Config, error) {

	cn := Config{}

	data, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config %w", err)
	}

	decoder := yaml.NewDecoder(data)
	decoder.KnownFields(true)
	err = decoder.Decode(&cn)

	if err != nil {
		return nil, fmt.Errorf("unable to parse config %w", err)
	}

	return &cn, nil
}
