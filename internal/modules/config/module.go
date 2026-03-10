package config

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Module struct{}

func NewModule() *Module {
	env := os.Getenv("APP_ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs

	if env == "development" {
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05.000"}
		log.Logger = log.Output(output)
	}

	log.Debug().Msg("Config Module initialized successfully")

	return &Module{}
}
