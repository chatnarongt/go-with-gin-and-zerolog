package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/middleware"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/logger"
)

type ModuleOptions struct {
	// Logger is the logger module options.
	Logger logger.ModuleOptions
	// Config is the configuration module options.
	Config config.ModuleOptions
	// Imports is a list of additional modules to import into the application.
	//
	// These modules will be registered and initialized after the core modules (logger and config).
	Imports []internal.Module
}

type Module struct {
	config  *config.Module
	logger  *logger.Module
	imports []internal.Module
}

var _ internal.Module = (*Module)(nil)

func NewModule(options ModuleOptions) *Module {
	return &Module{
		config:  config.NewModule(options.Config),
		logger:  logger.NewModule(options.Logger),
		imports: options.Imports,
	}
}

func (m *Module) coreModules() []internal.Module {
	return []internal.Module{m.logger, m.config}
}

func (m *Module) allModules() []internal.Module {
	return append(m.coreModules(), m.imports...)
}

func (m *Module) Start() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return m.StartContext(ctx)
}

func (m *Module) StartContext(ctx context.Context) error {
	injector := do.New()
	log := zerolog.Nop()

	if err := m.registerCore(injector); err != nil {
		log.Error().Err(err).Msg("Register application modules")
		return err
	}

	config, err := do.Invoke[*config.Config](injector)
	if err != nil {
		log.Error().Err(err).Msg("Load application config")
		return err
	}
	log = *do.MustInvoke[*zerolog.Logger](injector)

	if !config.Application.DebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(
		gin.Recovery(),
		middleware.RequestID(),
		middleware.AccessLog(func() *zerolog.Logger {
			return do.MustInvoke[*zerolog.Logger](injector)
		}),
	)

	if err := m.registerImports(injector, router); err != nil {
		log.Error().Err(err).Msg("Register application modules")
		return err
	}
	if err := m.initializeModules(); err != nil {
		log.Error().Err(err).Msg("Initialize application modules")
		return err
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Application.Port),
		Handler: router,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		log.Error().Err(err).Msg("HTTP server stopped unexpectedly")
		return err
	case <-ctx.Done():
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if config.Application.DebugMode {
		// Print a newline to separate the shutdown log from any previous logs in debug mode
		println("")
	}

	if err := m.shutdown(shutdownContext, server, injector); err != nil {
		log.Error().Err(err).Msg("Application shutdown failed")
		return err
	}
	log.Info().Msg("Application shut down successfully")
	return nil
}

func (m *Module) registerCore(i do.Injector) error {
	for _, module := range m.coreModules() {
		if err := module.Register(i, nil); err != nil {
			return err
		}
	}
	return nil
}

func (m *Module) registerImports(i do.Injector, router *gin.Engine) error {
	for _, module := range m.imports {
		if err := module.Register(i, router); err != nil {
			return err
		}
	}
	return nil
}

func (m *Module) initializeModules() error {
	for _, module := range m.allModules() {
		if lifecycle, ok := module.(internal.OnModuleInit); ok {
			if err := lifecycle.OnModuleInit(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Module) Register(i do.Injector, router *gin.Engine) error {
	if err := m.registerCore(i); err != nil {
		return err
	}
	if err := m.registerImports(i, router); err != nil {
		return err
	}
	return m.initializeModules()
}

func (m *Module) destroy(ctx context.Context) error {
	var destroyErrors []error

	for index := len(m.imports) - 1; index >= 0; index-- {
		lifecycle, ok := m.imports[index].(internal.OnModuleDestroy)
		if !ok {
			continue
		}

		errCh := make(chan error, 1)
		go func() {
			errCh <- lifecycle.OnModuleDestroy(ctx)
		}()

		select {
		case err := <-errCh:
			if err != nil {
				destroyErrors = append(destroyErrors, err)
			}
		case <-ctx.Done():
			return errors.Join(append(destroyErrors, ctx.Err())...)
		}
	}

	return errors.Join(destroyErrors...)
}

func (m *Module) shutdown(ctx context.Context, server *http.Server, injector do.Injector) error {
	serverErr := server.Shutdown(ctx)
	injectorReport := injector.ShutdownWithContext(ctx)
	var injectorErr error
	if len(injectorReport.Errors) > 0 {
		injectorErr = injectorReport
	}
	moduleErr := m.destroy(ctx)
	return errors.Join(serverErr, injectorErr, moduleErr)
}
