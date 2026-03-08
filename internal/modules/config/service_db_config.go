package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type DBConfig struct {
	Host     string `validate:"max=255"`
	User     string `validate:"min=1,max=115"`
	Password string `validate:"min=8,max=128"`
	Name     string `validate:"max=128"`

	Port                          int `validate:"number,min=1,max=65535"`
	MaxIdleConnectionSize         int `validate:"min=0"`
	MaxConnectionSize             int `validate:"min=0"`
	MaxConnectionIdleTimeInSecond int `validate:"min=0"`
	MaxConnectionLifeTimeInSecond int `validate:"min=0"`
}

func (m *Module) LoadDBConfig() *DBConfig {
	config := &DBConfig{
		Host:     getEnvAsString("DB_HOST", "localhost"),
		User:     getEnvAsString("DB_USER", "sa"),
		Password: getEnvAsString("DB_PASSWORD", ""),
		Name:     getEnvAsString("DB_NAME", "master"),

		Port:                          getEnvAsInt("DB_PORT", 1433),
		MaxIdleConnectionSize:         getEnvAsInt("DB_MAX_IDLE_CONNECTION_SIZE", 10),
		MaxConnectionSize:             getEnvAsInt("DB_MAX_CONNECTION_SIZE", 100),
		MaxConnectionIdleTimeInSecond: getEnvAsInt("DB_MAX_CONNECTION_IDLE_TIME", 300),
		MaxConnectionLifeTimeInSecond: getEnvAsInt("DB_MAX_CONNECTION_LIFE_TIME", 3600),
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(config); err != nil {
		log.Fatal().Err(err).Msg("Invalid database configuration")
	}

	return config
}
