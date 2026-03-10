package main

import (
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/application"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/database"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/health"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/swagger"
)

// @title Go with Gin and Zerolog
// @version 1.0.0
// @description This is a sample server.
func main() {
	cfg := config.NewModule()

	acf := cfg.LoadAppConfig()

	dbm := database.NewModule(cfg)

	swm := swagger.NewModule(acf)

	htm := health.NewModule(dbm.DB)

	app := application.NewModule(acf)

	app.MapRoutes(swm)

	app.MapAPIRoutes(htm)

	app.OnAfterShutdown(
		dbm.Cleanup,
	)

	app.ListenAndServe()
}
