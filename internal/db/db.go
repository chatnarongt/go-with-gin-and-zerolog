package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/config"
	_ "github.com/microsoft/go-mssqldb"
)

// NewDB creates and validates a new database connection using the provided config.
// Returns the DB instance, a cleanup function to close it, and any error.
// Wire will automatically call the cleanup function on shutdown.
func NewDB(cfg *config.DBConfig) (*sql.DB, func(), error) {
	connString := fmt.Sprintf(
		"Server=%s;Port=%d;Database=%s;User ID=%s;Password=%s;",
		cfg.Host, cfg.Port, cfg.Name, cfg.User, cfg.Password,
	)

	db, err := sql.Open("sqlserver", connString)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open db connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to ping db: %w", err)
	}

	db.SetMaxIdleConns(cfg.MaxIdleConnectionSize)
	db.SetMaxOpenConns(cfg.MaxConnectionSize)
	db.SetConnMaxIdleTime(time.Duration(cfg.MaxConnectionIdleTimeInSecond) * time.Second)
	db.SetConnMaxLifetime(time.Duration(cfg.MaxConnectionLifeTimeInSecond) * time.Second)

	cleanup := func() {
		db.Close()
	}

	return db, cleanup, nil
}
