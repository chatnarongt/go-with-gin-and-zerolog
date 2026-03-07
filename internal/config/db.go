package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

type DBConfig struct {
	Host                          string `validate:"required"`
	Port                          int    `validate:"required,gt=0"`
	User                          string `validate:"required"`
	Password                      string `validate:"required"`
	Name                          string `validate:"required"`
	MaxIdleConnectionSize         int    `validate:"gte=0"`
	MaxConnectionSize             int    `validate:"gte=0"`
	MaxConnectionIdleTimeInSecond int    `validate:"gte=0"`
	MaxConnectionLifeTimeInSecond int    `validate:"gte=0"`
}

func LoadDBConfig() (*DBConfig, error) {
	cfg := &DBConfig{
		Host:                          getEnvOrDefault("DB_HOST", "localhost"),
		Port:                          getEnvAsIntOrDefault("DB_PORT", 1433),
		User:                          getEnvOrDefault("DB_USER", "sa"),
		Password:                      getEnvOrDefault("DB_PASSWORD", ""),
		Name:                          getEnvOrDefault("DB_NAME", "master"),
		MaxIdleConnectionSize:         getEnvAsIntOrDefault("DB_CONN_IDLE_SIZE", 1),
		MaxConnectionSize:             getEnvAsIntOrDefault("DB_CONN_MAX_SIZE", 1),
		MaxConnectionIdleTimeInSecond: getEnvAsIntOrDefault("DB_CONN_IDLE_TIME_SEC", 60),
		MaxConnectionLifeTimeInSecond: getEnvAsIntOrDefault("DB_CONN_LIFE_TIME_SEC", 60),
	}

	if err := validator.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("Invalid db config: %w", err)
	}

	return cfg, nil
}
