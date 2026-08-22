package database_test

import (
	"context"
	"os"
	"testing"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/database"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
)

func TestDatabaseModule_SQLite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	defer os.Remove("./data/test_main.db")
	defer os.Remove("./data/test_analytics.db")

	injector := do.New()
	nopLogger := zerolog.Nop()
	do.ProvideValue(injector, &nopLogger)

	cfg := &config.Config{
		Databases: config.DatabasesConfig{
			Main: config.DatabaseConnectionConfig{
				Driver:       "sqlite",
				DSN:          "file:./data/test_main.db",
				Required:     true,
				MaxOpenConns: 2,
				MaxIdleConns: 1,
			},
			Analytics: config.DatabaseConnectionConfig{
				Driver:       "sqlite",
				DSN:          "file:./data/test_analytics.db",
				Required:     true,
				ReadOnly:     true,
				MaxOpenConns: 2,
				MaxIdleConns: 1,
			},
			Logging: config.DatabaseConnectionConfig{
				Driver:   "mongodb",
				DSN:      "",
				Required: false,
			},
		},
	}
	do.ProvideValue(injector, cfg)

	dbModule := database.NewModule()
	if err := dbModule.Register(injector, nil); err != nil {
		t.Fatalf("register database module: %v", err)
	}

	db := do.MustInvoke[*database.Databases](injector)
	if db.Main == nil || db.Analytics == nil {
		t.Fatal("expected Main and Analytics DB connections to be non-nil")
	}

	if err := dbModule.OnModuleInit(); err != nil {
		t.Fatalf("initialize database module: %v", err)
	}

	// Verify Main connection write
	if _, err := db.Main.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY);"); err != nil {
		t.Fatalf("failed to exec on main db: %v", err)
	}

	// Verify Analytics read-only rejection on write
	if _, err := db.Analytics.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY);"); err == nil {
		t.Fatal("expected write operation on read-only analytics DB to fail")
	}

	if err := dbModule.OnModuleDestroy(context.Background()); err != nil {
		t.Fatalf("destroy database module: %v", err)
	}
}

func TestDatabaseModule_MongoMissingDatabaseFails(t *testing.T) {
	gin.SetMode(gin.TestMode)

	injector := do.New()
	nopLogger := zerolog.Nop()
	do.ProvideValue(injector, &nopLogger)

	cfg := &config.Config{
		Databases: config.DatabasesConfig{
			Main: config.DatabaseConnectionConfig{
				Driver: "sqlite",
				DSN:    "file:./data/test_main_fail.db",
			},
			Analytics: config.DatabaseConnectionConfig{
				Driver: "sqlite",
				DSN:    "file:./data/test_analytics_fail.db",
			},
			Logging: config.DatabaseConnectionConfig{
				Driver:   "mongodb",
				DSN:      "mongodb://localhost:27017",
				Required: true,
			},
		},
	}
	defer os.Remove("./data/test_main_fail.db")
	defer os.Remove("./data/test_analytics_fail.db")

	do.ProvideValue(injector, cfg)

	dbModule := database.NewModule()
	if err := dbModule.Register(injector, nil); err == nil {
		t.Fatal("expected register to fail when required MongoDB DSN has no database name")
	}
}
