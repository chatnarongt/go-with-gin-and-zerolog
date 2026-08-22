package main

import (
	_ "embed"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/application"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/database"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/probe"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/swagger"
)

//go:embed openapi.yaml
var openAPISpec []byte

func main() {
	app := application.NewModule(application.ModuleOptions{
		Config: config.ModuleOptions{
			EnvFiles: []string{".env", "cmd/api/.env"},
		},
		Imports: []internal.Module{
			database.NewModule(),
			swagger.NewModule(swagger.ModuleOptions{
				Spec:        openAPISpec,
				Title:       "Go with Gin and Zerolog API",
				Version:     "0.1.0",
				Description: "HTTP API documentation.",
			}),
			probe.NewModule(),
		},
	})

	app.Start()
}
