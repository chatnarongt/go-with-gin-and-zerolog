package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/rs/zerolog/log"
)

type Module struct {
	DB      *sql.DB
	Cleanup func()
}

func NewModule(config *config.Module) *Module {
	cfg := config.LoadDBConfig()

	u := &url.URL{
		Scheme: "sqlserver",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
	}
	q := u.Query()
	q.Set("Database", cfg.Name)
	u.RawQuery = q.Encode()

	db, err := sql.Open("sqlserver", u.String())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open db connection")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping db")
	}

	db.SetMaxIdleConns(cfg.MaxIdleConnectionSize)
	db.SetMaxOpenConns(cfg.MaxConnectionSize)
	db.SetConnMaxIdleTime(time.Duration(cfg.MaxConnectionIdleTimeInSecond) * time.Second)
	db.SetConnMaxLifetime(time.Duration(cfg.MaxConnectionLifeTimeInSecond) * time.Second)

	log.Debug().Msg("Database Module initialized successfully")

	return &Module{
		DB: db,
		Cleanup: func() {
			if err := db.Close(); err != nil {
				log.Error().Err(err).Msg("Failed to close db connection")
			}
			log.Debug().Msg("Database connection closed")
		},
	}
}
