package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DatabaseURL  string `env:"DATABASE_URL,required"`
	BaseShortURL string `env:"BASE_SHORT_URL,required"`
	Port         string `env:"SERVER_PORT" envDefault:"8080"`
	CORSOrigin   string `env:"CORS_ORIGIN" envDefault:"http://localhost:5173"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}

	return &cfg, nil
}
