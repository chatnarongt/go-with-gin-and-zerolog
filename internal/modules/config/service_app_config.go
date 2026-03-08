package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type AppConfig struct {
	Environment   string `validate:"oneof=development staging production"`
	Port          int    `validate:"number,min=1,max=65535"`
	EnableSwagger bool   `validate:"boolean"`
}

func (m *Module) LoadAppConfig() *AppConfig {
	config := &AppConfig{
		Environment:   getEnvAsString("APP_ENVIRONMENT", "development"),
		Port:          getEnvAsInt("APP_PORT", 8080),
		EnableSwagger: getEnvAsBool("APP_ENABLE_SWAGGER", false),
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(config); err != nil {
		log.Fatal().Err(err).Msg("Invalid app configuration")
	}

	return config
}
