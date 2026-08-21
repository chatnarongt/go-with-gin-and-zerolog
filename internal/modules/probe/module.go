package probe

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
)

type Module struct {
	service *Service
	logger  *zerolog.Logger
}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Register(i do.Injector, router *gin.Engine) error {
	m.logger = do.MustInvoke[*zerolog.Logger](i)

	do.Provide(i, NewService)
	do.Provide(i, NewController)

	service, err := do.Invoke[*Service](i)
	if err != nil {
		return err
	}
	m.service = service

	controller, err := do.Invoke[*Controller](i)
	if err != nil {
		return err
	}
	controller.RegisterRoutes(router)
	return nil
}

func (m *Module) OnModuleInit() error {
	m.logger.Info().Msg("Probe module initialized")
	return m.service.OnModuleInit()
}
