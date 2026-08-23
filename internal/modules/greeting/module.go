package greeting

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal"
)

type Module struct {
	logger *zerolog.Logger
}

var _ internal.Module = (*Module)(nil)

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Register(i do.Injector, router *gin.Engine) error {
	m.logger = do.MustInvoke[*zerolog.Logger](i)

	do.Provide(i, NewJob)

	job, err := do.Invoke[*Job](i)
	if err != nil {
		return err
	}

	if registry, err := do.Invoke[internal.JobRegistry](i); err == nil && registry != nil {
		job.RegisterJobs(registry)
	}

	return nil
}

func (m *Module) OnModuleInit() error {
	m.logger.Info().Msg("Greeting module initialized")
	return nil
}
