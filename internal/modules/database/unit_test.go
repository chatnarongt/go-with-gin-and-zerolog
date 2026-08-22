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
				MaxOpenConns: 2,
				MaxIdleConns: 1,
			},
			Analytics: config.DatabaseConnectionConfig{
				Driver:       "sqlite",
				DSN:          "file:./data/test_analytics.db",
				ReadOnly:     true,
				MaxOpenConns: 2,
				MaxIdleConns: 1,
			},
		},
	}
	do.ProvideValue(injector, cfg)

	dbModule := database.NewModule()
	if err := dbModule.Register(injector, nil); err != nil {
		t.Fatalf("register database module: %v", err)
	}

	dbs := do.MustInvoke[*database.Databases](injector)
	if dbs.Main == nil || dbs.Analytics == nil {
		t.Fatal("expected Main and Analytics DB connections to be non-nil")
	}

	if err := dbModule.OnModuleInit(); err != nil {
		t.Fatalf("initialize database module: %v", err)
	}

	// Verify Main connection write
	if _, err := dbs.Main.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY);"); err != nil {
		t.Fatalf("failed to exec on main db: %v", err)
	}

	// Verify Analytics read-only rejection on write
	if _, err := dbs.Analytics.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY);"); err == nil {
		t.Fatal("expected write operation on read-only analytics DB to fail")
	}

	if err := dbModule.OnModuleDestroy(context.Background()); err != nil {
		t.Fatalf("destroy database module: %v", err)
	}
}
