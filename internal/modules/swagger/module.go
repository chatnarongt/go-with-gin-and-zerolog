package swagger

import "github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"

type Module struct {
	config *config.Module
}

func NewModule(config *config.Module) *Module {
	return &Module{
		config: config,
	}
}
