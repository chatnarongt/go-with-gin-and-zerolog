package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
)

type AppConfig struct {
	Environment string `validate:"required,oneof=local development staging production"`
	Port        string `validate:"required,numeric"`
}

func LoadAppConfig() (*AppConfig, error) {
	cfg := &AppConfig{
		Environment: os.Getenv("APP_ENVIRONMENT"),
		Port:        os.Getenv("APP_PORT"),
	}

	if err := validator.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("Invalid app config: %w", err)
	}

	return cfg, nil
}
