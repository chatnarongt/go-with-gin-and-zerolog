package main

import (
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/application"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/database"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/health"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/schedule"
)

func main() {
	cfg := config.NewModule()

	acf := cfg.LoadAppConfig()

	dbm := database.NewModule(cfg)

	htm := health.NewModule(dbm.DB)

	scm := schedule.NewModule(dbm)

	app := application.NewModule(acf)

	app.MapAPIRoutes(htm)

	app.OnAfterShutdown(
		scm.Cleanup,
		dbm.Cleanup,
	)

	app.ListenAndServe()
}
