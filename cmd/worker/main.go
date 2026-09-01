package main

import (
	"fmt"
	"os"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/database"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/greeting"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/probe"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/transaction"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/worker"
)

func main() {
	app := worker.NewModule(worker.ModuleOptions{
		Config: config.ModuleOptions{
			EnvFiles: []string{".env"},
		},
		Imports: []internal.Module{
			database.NewModule(),
			probe.NewModule(),
			greeting.NewModule(),
			transaction.NewModule(),
		},
	})

	if err := app.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Worker error: %v\n", err)
		os.Exit(1)
	}
}
