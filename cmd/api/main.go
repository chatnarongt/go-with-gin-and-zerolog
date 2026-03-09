package main

import (
	"os"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/application"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/database"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/health"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/swagger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// @title Go with Gin and Zerolog
// @version 1.0.0
// @description This is a sample server.
func main() {
	cfg := config.NewModule()

	dbm := database.NewModule(cfg)

	swm := swagger.NewModule(cfg)

	htm := health.NewModule(dbm.DB)

	app := application.NewModule(cfg)

	app.MapRoutes(swm)

	app.MapAPIRoutes(htm)

	app.OnBeforeShutdown(
		dbm.Cleanup,
	)

	app.ListenAndServe()
}

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05.000"}
	log.Logger = log.Output(output)
}
