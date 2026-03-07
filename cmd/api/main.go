package main

import (
	"os"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/config"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/db"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/health"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// @title Go with Gin and Zerolog
// @version 1.0.0
// @description This is a sample server.
// @BasePath /api
func main() {
	// Setup logger
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}
	log.Logger = log.Output(output)

	// -- 1. Load Configurations --
	appConfig, err := config.LoadAppConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load app config")
	}

	dbConfig, err := config.LoadDBConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load db config")
	}

	// -- 2. Create Core Infrastructure --
	dbClient, cleanup, err := db.NewDB(dbConfig)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer cleanup()

	s := server.NewAPIServer(appConfig)

	api := s.Router.Group("/api")

	// -- 3. Initialize Modules & Routes --
	health.NewModule(dbClient).Register(api)

	// -- 4. Start Server --
	if err := s.Start(); err != nil {
		log.Fatal().Msgf("server error: %v", err)
	}
}
