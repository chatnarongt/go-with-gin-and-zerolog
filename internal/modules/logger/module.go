package logger

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
)

type ModuleOptions struct {
	// Hooks is a list of zerolog hooks to apply to the logger.
	//
	// Hooks can be used to add additional fields to log entries, or to modify the log level.
	Hooks []zerolog.Hook
}

type Module struct {
	hooks  []zerolog.Hook
	logger *zerolog.Logger
}

func NewModule(options ModuleOptions) *Module {
	return &Module{hooks: options.Hooks}
}

func (m *Module) Register(i do.Injector, _ *gin.Engine) error {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	do.Provide(i, func(i do.Injector) (*zerolog.Logger, error) {
		config := do.MustInvoke[*config.Config](i)

		var log zerolog.Logger

		if config.Application.DebugMode {
			cw := zerolog.ConsoleWriter{
				Out:        os.Stdout,
				TimeFormat: "2006-01-02 15:04:05.000000000",
			}
			log = zerolog.New(cw)
		} else {
			log = zerolog.New(os.Stdout)
		}

		log = log.Level(config.Application.LogLevel).Hook(m.hooks...).With().Timestamp().Logger()
		m.logger = &log

		return &log, nil
	})
	return nil
}

func (m *Module) OnModuleInit() error {
	m.logger.Info().Msg("Logger module initialized")
	return nil
}
