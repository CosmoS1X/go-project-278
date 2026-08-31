package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DatabaseURL  string `env:"DATABASE_URL,required"`
	BaseShortURL string `env:"BASE_SHORT_URL,required"`
	Port         string `env:"PORT" envDefault:"8080"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}

	return &cfg, nil
}
