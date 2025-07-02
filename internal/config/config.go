package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Logger   LoggerConfig   `yaml:"logger"`
}

type DatabaseConfig struct {
	Host     string `yaml:"localhost"`
	Port     int    `yaml:"5432"`
	User     string `yaml:"admino"`
	Password string `yaml:"admino"`
	Name     string `yaml:"pandora_logs"`
	SSLMode  string `yaml:"ssl-mode"`
}

type LoggerConfig struct {
	Level string `yaml:"level"`
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
