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

	_ "github.com/chatnarongt/go-with-gin-and-zerolog/docs"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/middleware"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Module struct {
	config            *config.Module
	appConfig         *config.AppConfig
	server            *http.Server
	onBeforeShutdowns []func()
	Router            *gin.RouterGroup
}

func NewModule(config *config.Module) *Module {
	appConfig := config.LoadAppConfig()

	if appConfig.Environment != "development" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		fmt.Println() // Just to add a newline after the server startup log for better readability in development mode
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandler())

	router.NoRoute(notFound)

	if appConfig.EnableSwagger {
		router.GET("/swagger", func(c *gin.Context) {
			c.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
		})
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	log.Debug().Msg("Application Module initialized successfully")

	return &Module{
		config:    config,
		appConfig: appConfig,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", appConfig.Port),
			Handler: router,
		},
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
		fmt.Println() // Just to add a newline after the server startup log for better readability in development mode
	}
	log.Info().Msg("Server started successfully")
	log.Info().Msgf("Listening on http://localhost%s", m.server.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Fatal().Msgf("Server error: %v", err)
	case sig := <-quit:
		fmt.Print("\n")
		log.Info().Msgf("Received signal %s, shutting down gracefully...", sig)
	}

	m.shutdown()
}

func (m *Module) shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server shutdown error")
	}

	for _, f := range m.onBeforeShutdowns {
		f()
	}

	log.Info().Msg("Server shut down gracefully")
}
