package config

import "github.com/rs/zerolog/log"

type Module struct{}

func NewModule() *Module {
	log.Debug().Msg("Config Module initialized successfully")

	return &Module{}
}
