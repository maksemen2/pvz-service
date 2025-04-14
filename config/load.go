package config

import (
	"github.com/caarlos0/env/v6"
)

// Load загружает конфиг из переменных окружения
func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
