package server

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
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/config"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type APIServer struct {
	server    *http.Server
	appConfig *config.AppConfig
	Router    *gin.Engine // เปิดเผย Router ให้ modules สามารถ RegisterRoutes ได้
}

// NewAPIServer creates a new empty server. Routes will be registered from outside.
func NewAPIServer(appConfig *config.AppConfig) *APIServer {
	if appConfig.Environment != "local" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.ErrorHandler())

	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
	})
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return &APIServer{
		appConfig: appConfig,
		Router:    router,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%s", appConfig.Port),
			Handler: router,
		},
	}
}

// Start runs the HTTP server in a goroutine and blocks until an OS
// interrupt or termination signal is received, then gracefully shuts down.
func (s *APIServer) Start() error {
	errCh := make(chan error, 1)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	if s.appConfig.Environment == "local" {
		fmt.Println()
	}
	log.Info().Msg("Server started successfully")
	log.Info().Msgf("Listening on http://localhost%s", s.server.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		fmt.Print("`\n")
		log.Info().Msgf("Received signal %s, shutting down gracefully...", sig)
	}

	return s.Stop()
}

// Stop gracefully shuts down the server with a 10-second timeout.
func (s *APIServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	log.Info().Msg("Server stopped gracefully")
	return nil
}
