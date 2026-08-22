package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	_ "modernc.org/sqlite"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
)

const startupPingTimeout = 5 * time.Second

type Databases struct {
	Main      *sql.DB
	Analytics *sql.DB
	Logging   *mongo.Database
}

type Module struct {
	db     *Databases
	cfg    *config.DatabasesConfig
	logger *zerolog.Logger
}

var _ internal.Module = (*Module)(nil)

func NewModule() *Module {
	return &Module{}
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

func openMongo(cfg config.DatabaseConnectionConfig) (*mongo.Client, *mongo.Database, error) {
	dsn := cfg.DSN
	if cfg.ReadOnly {
		dsn = enrichReadOnlyDSN("mongodb", dsn)
	}

	clientOpts := options.Client().ApplyURI(dsn)
	if cfg.MaxOpenConns > 0 {
		maxPool := uint64(cfg.MaxOpenConns)
		clientOpts.SetMaxPoolSize(maxPool)
	}
	if cfg.MaxIdleConns > 0 {
		minPool := uint64(cfg.MaxIdleConns)
		clientOpts.SetMinPoolSize(minPool)
	}
	if cfg.ConnMaxIdleTime > 0 {
		clientOpts.SetMaxConnIdleTime(cfg.ConnMaxIdleTime)
	}
	if cfg.ReadOnly {
		clientOpts.SetReadPreference(readpref.Secondary())
	}

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("connect mongodb: %w", err)
	}

	dbName, err := parseMongoDatabaseName(cfg.DSN)
	if err != nil {
		return nil, nil, err
	}
	return client, client.Database(dbName), nil
}

func (m *Module) Register(i do.Injector, _ *gin.Engine) error {
	applicationConfig := do.MustInvoke[*config.Config](i)
	m.logger = do.MustInvoke[*zerolog.Logger](i)
	m.cfg = &applicationConfig.Databases

	mainDB, err := openDatabase(applicationConfig.Databases.Main)
	if err != nil {
		return fmt.Errorf("open main database: %w", err)
	}

	analyticsDB, err := openDatabase(applicationConfig.Databases.Analytics)
	if err != nil {
		_ = mainDB.Close()
		return fmt.Errorf("open analytics database: %w", err)
	}

	var loggingDB *mongo.Database
	if applicationConfig.Databases.Logging.DSN != "" {
		client, db, err := openMongo(applicationConfig.Databases.Logging)
		if err != nil {
			if applicationConfig.Databases.Logging.Required {
				_ = mainDB.Close()
				_ = analyticsDB.Close()
				return fmt.Errorf("open logging database: %w", err)
			}
			m.logger.Warn().Err(err).Msg("Failed to initialize optional logging database")
		} else {
			loggingDB = db
			_ = client
		}
	}

	m.db = &Databases{
		Main:      mainDB,
		Analytics: analyticsDB,
		Logging:   loggingDB,
	}

	do.ProvideValue(i, m.db)
	do.ProvideNamedValue(i, "main", m.db.Main)
	do.ProvideNamedValue(i, "analytics", m.db.Analytics)
	if loggingDB != nil {
		do.ProvideValue(i, loggingDB.Client())
		do.ProvideValue(i, loggingDB)
		do.ProvideNamedValue(i, "logging", loggingDB)
	}
	do.ProvideValue(i, map[string]*sql.DB{
		"main":      m.db.Main,
		"analytics": m.db.Analytics,
	})
	return nil
}

func (m *Module) OnModuleInit() error {
	if m.db == nil {
		return fmt.Errorf("initialize database module: databases not registered")
	}

	ctx, cancel := context.WithTimeout(context.Background(), startupPingTimeout)
	defer cancel()

	if m.cfg.Main.Required {
		if err := m.db.Main.PingContext(ctx); err != nil {
			return fmt.Errorf("ping main database: %w", err)
		}
	}
	if m.cfg.Analytics.Required {
		if err := m.db.Analytics.PingContext(ctx); err != nil {
			return fmt.Errorf("ping analytics database: %w", err)
		}
	}

	if m.db.Logging != nil {
		if err := m.db.Logging.Client().Ping(ctx, nil); err != nil {
			if m.cfg.Logging.Required {
				return fmt.Errorf("ping logging database: %w", err)
			}
			m.logger.Warn().Err(err).Msg("Optional logging database ping failed at startup")
		}
	} else if m.cfg.Logging.Required {
		return fmt.Errorf("ping logging database: required logging database not initialized")
	}

	m.logger.Info().Msg("Database module initialized")
	return nil
}

func (m *Module) OnModuleDestroy(ctx context.Context) error {
	if m.db == nil {
		return nil
	}

	var closeErrors []error
	if m.db.Main != nil {
		if err := m.db.Main.Close(); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("close main database: %w", err))
		}
	}
	if m.db.Analytics != nil {
		if err := m.db.Analytics.Close(); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("close analytics database: %w", err))
		}
	}
	if m.db.Logging != nil {
		if err := m.db.Logging.Client().Disconnect(ctx); err != nil {
			closeErrors = append(closeErrors, fmt.Errorf("close logging database: %w", err))
		}
	}

	if len(closeErrors) > 0 {
		return errors.Join(closeErrors...)
	}

	m.logger.Info().Msg("Successfully closed all database connections")
	return nil
}
