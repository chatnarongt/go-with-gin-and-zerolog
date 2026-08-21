package main

import (
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/application"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/database"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/probe"
)

func main() {
	app := application.NewModule(application.ModuleOptions{
		Config: config.ModuleOptions{
			EnvFiles: []string{".env", "cmd/api/.env"},
		},
		Imports: []internal.Module{
			database.NewModule(),
			probe.NewModule(),
		},
	})

	app.Start()
}
