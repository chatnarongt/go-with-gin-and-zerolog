package main

import (
	"os"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/application"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/database"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/schedule"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg := config.NewModule()

	db := database.NewModule(cfg)

	sm := schedule.NewModule(db)

	app := application.NewModule(cfg)

	app.OnBeforeShutdown(
		sm.Cleanup,
		db.Cleanup,
	)

	app.ListenAndServe()
}

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05.000"}
	log.Logger = log.Output(output)
}
