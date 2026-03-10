package swagger

import "github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"

type Module struct {
	appConfig *config.AppConfig
}

func NewModule(appConfig *config.AppConfig) *Module {
	return &Module{
		appConfig: appConfig,
	}
}
