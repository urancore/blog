package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

func Load(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("config path is not empty")
	}

	var config Config
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	return &config, nil
}
