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

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/middleware"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Module struct {
	config            *config.Module
	appConfig         *config.AppConfig
	server            *http.Server
	onBeforeShutdowns []func()
	engine            *gin.Engine
	Router            *gin.RouterGroup
}

func NewModule(config *config.Module) *Module {
	appConfig := config.LoadAppConfig()

	if appConfig.Environment != "development" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		// Just to add a newline after the server startup log for better readability in development mode
		fmt.Println()
		defer fmt.Println() // Print a newline before function exit
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandler())

	router.NoRoute(notFound)

	log.Debug().Msg("Application Module initialized successfully")

	return &Module{
		config:    config,
		appConfig: appConfig,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", appConfig.Port),
			Handler: router,
		},
		engine: router,
		Router: router.Group("/api"),
	}
}

func (m *Module) ListenAndServe() {
	errCh := make(chan error, 1)

	go func() {
		if err := m.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	if m.appConfig.Environment == "development" {
		// Just to add a newline after the server startup log for better readability in development mode
		fmt.Println()
	}
	log.Info().Msg("Server started successfully")
	log.Info().Msgf("Listening on http://localhost%s", m.server.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Fatal().Err(err).Msgf("Server error")
	case sig := <-quit:
		fmt.Print("\n")
		log.Info().Msgf("Received signal %s, shutting down...", sig)
	}

	m.shutdown()
}

func (m *Module) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server shut down error")
	}

	for _, f := range m.onBeforeShutdowns {
		f()
	}

	log.Info().Msg("Server shut down gracefully")
}
