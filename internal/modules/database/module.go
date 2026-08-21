package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	_ "modernc.org/sqlite"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
)

const startupPingTimeout = 5 * time.Second

type Module struct {
	db     *sql.DB
	logger *zerolog.Logger
}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Register(i do.Injector, _ *gin.Engine) error {
	applicationConfig := do.MustInvoke[*config.Config](i)
	m.logger = do.MustInvoke[*zerolog.Logger](i)
	if err := ensureDatabaseDirectory(applicationConfig.Database.DSN); err != nil {
		return err
	}

	db, err := sql.Open("sqlite", applicationConfig.Database.DSN)
	if err != nil {
		return fmt.Errorf("open sqlite database: %w", err)
	}

	db.SetMaxOpenConns(applicationConfig.Database.MaxOpenConns)
	db.SetMaxIdleConns(applicationConfig.Database.MaxIdleConns)
	db.SetConnMaxLifetime(applicationConfig.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(applicationConfig.Database.ConnMaxIdleTime)
	m.db = db
	do.ProvideValue(i, db)
	return nil
}

func ensureDatabaseDirectory(dsn string) error {
	if !strings.HasPrefix(dsn, "file:") {
		return nil
	}

	// Strip the leading "file:" so any path after it is handled portably;
	// url.Parse treats such DSNs as opaque and returns no path.
	path := strings.TrimPrefix(dsn, "file:")
	if slashIndex := strings.IndexByte(path, '?'); slashIndex >= 0 {
		path = path[:slashIndex]
	}
	if path == "" || strings.HasPrefix(path, ":memory:") {
		return nil
	}
	directory := filepath.Dir(path)
	if directory == "." {
		return nil
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create sqlite database directory %q: %w", directory, err)
	}
	return nil
}

func (m *Module) OnModuleInit() error {
	if m.db == nil {
		return fmt.Errorf("initialize sqlite database: database is not registered")
	}

	ctx, cancel := context.WithTimeout(context.Background(), startupPingTimeout)
	defer cancel()
	if err := m.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping sqlite database: %w", err)
	}

	m.logger.Info().Msg("Database module initialized")
	return nil
}

func (m *Module) OnModuleDestroy(context.Context) error {
	if m.db == nil {
		return nil
	}

	if err := m.db.Close(); err != nil {
		return fmt.Errorf("close sqlite database: %w", err)
	}

	m.logger.Info().Msg("Successfully closed sqlite database")
	return nil
}
