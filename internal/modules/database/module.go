package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	_ "modernc.org/sqlite"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
)

const startupPingTimeout = 5 * time.Second

type Databases struct {
	Main      *sql.DB
	Analytics *sql.DB
}

type Module struct {
	dbs    *Databases
	logger *zerolog.Logger
}

var _ internal.Module = (*Module)(nil)

func NewModule() *Module {
	return &Module{}
}

func normalizeDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "postgres", "pgx":
		return "pgx"
	case "sqlserver", "mssql":
		return "sqlserver"
	case "sqlite":
		return "sqlite"
	default:
		return driver
	}
}

func enrichReadOnlyDSN(driver, dsn string) string {
	switch driver {
	case "sqlite":
		if strings.Contains(dsn, "mode=ro") {
			return dsn
		}
		separator := "?"
		if strings.Contains(dsn, "?") {
			separator = "&"
		}
		return dsn + separator + "mode=ro"
	case "pgx":
		if strings.Contains(dsn, "default_transaction_read_only") {
			return dsn
		}
		separator := "?"
		if strings.Contains(dsn, "?") {
			separator = "&"
		}
		return dsn + separator + "options=-c%20default_transaction_read_only%3Don"
	case "sqlserver":
		if strings.Contains(strings.ToLower(dsn), "applicationintent") || strings.Contains(strings.ToLower(dsn), "app intent") {
			return dsn
		}
		if parsed, err := url.Parse(dsn); err == nil && parsed.Scheme == "sqlserver" {
			query := parsed.Query()
			query.Set("app intent", "ReadOnly")
			parsed.RawQuery = query.Encode()
			return parsed.String()
		}
		separator := ";"
		if strings.HasSuffix(dsn, ";") {
			separator = ""
		}
		return dsn + separator + "ApplicationIntent=ReadOnly"
	default:
		return dsn
	}
}

func openDatabase(cfg config.DatabaseConnectionConfig) (*sql.DB, error) {
	driver := normalizeDriver(cfg.Driver)
	dsn := cfg.DSN
	if cfg.ReadOnly {
		dsn = enrichReadOnlyDSN(driver, dsn)
	}

	if driver == "sqlite" {
		if err := ensureDatabaseDirectory(dsn); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s database: %w", driver, err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	return db, nil
}

func (m *Module) Register(i do.Injector, _ *gin.Engine) error {
	applicationConfig := do.MustInvoke[*config.Config](i)
	m.logger = do.MustInvoke[*zerolog.Logger](i)

	mainDB, err := openDatabase(applicationConfig.Databases.Main)
	if err != nil {
		return fmt.Errorf("open main database: %w", err)
	}

	analyticsDB, err := openDatabase(applicationConfig.Databases.Analytics)
	if err != nil {
		_ = mainDB.Close()
		return fmt.Errorf("open analytics database: %w", err)
	}

	m.dbs = &Databases{
		Main:      mainDB,
		Analytics: analyticsDB,
	}

	do.ProvideValue(i, m.dbs)
	do.ProvideNamedValue(i, "main", m.dbs.Main)
	do.ProvideNamedValue(i, "analytics", m.dbs.Analytics)
	do.ProvideValue(i, map[string]*sql.DB{
		"main":      m.dbs.Main,
		"analytics": m.dbs.Analytics,
	})
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

	// If read-only mode, SQLite will fail to open if file does not exist yet.
	// Ensure an empty file exists so it can be opened.
	if strings.Contains(dsn, "mode=ro") {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
			if err != nil {
				return fmt.Errorf("create read-only sqlite database file %q: %w", path, err)
			}
			_ = file.Close()
		}
	}

	return nil
}

func (m *Module) OnModuleInit() error {
	if m.dbs == nil {
		return fmt.Errorf("initialize database module: databases not registered")
	}

	ctx, cancel := context.WithTimeout(context.Background(), startupPingTimeout)
	defer cancel()

	if err := m.dbs.Main.PingContext(ctx); err != nil {
		return fmt.Errorf("ping main database: %w", err)
	}
	if err := m.dbs.Analytics.PingContext(ctx); err != nil {
		return fmt.Errorf("ping analytics database: %w", err)
	}

	m.logger.Info().Msg("Database module initialized")
	return nil
}

func (m *Module) OnModuleDestroy(context.Context) error {
	if m.dbs == nil {
		return nil
	}

	var closeErrors []error
	if m.dbs.Main != nil {
		if err := m.dbs.Main.Close(); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("close main database: %w", err))
		}
	}
	if m.dbs.Analytics != nil {
		if err := m.dbs.Analytics.Close(); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("close analytics database: %w", err))
		}
	}

	if len(closeErrors) > 0 {
		return errors.Join(closeErrors...)
	}

	m.logger.Info().Msg("Successfully closed all database connections")
	return nil
}
