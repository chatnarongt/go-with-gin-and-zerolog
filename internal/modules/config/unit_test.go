package config_test

import (
	"testing"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
)

func TestConfigModule_Databases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	injector := do.New()
	nopLogger := zerolog.Nop()
	do.ProvideValue(injector, &nopLogger)

	cfgModule := config.NewModule(config.ModuleOptions{})
	if err := cfgModule.Register(injector, nil); err != nil {
		t.Fatalf("register config module: %v", err)
	}

	cfg := do.MustInvoke[*config.Config](injector)
	if cfg.Databases.Main.Driver != "sqlite" {
		t.Errorf("expected default driver sqlite, got %s", cfg.Databases.Main.Driver)
	}
	if cfg.Databases.Main.DSN != "file:./data/app.db" {
		t.Errorf("expected default dsn file:./data/app.db, got %s", cfg.Databases.Main.DSN)
	}
	if cfg.Databases.Analytics.Driver != "sqlite" {
		t.Errorf("expected analytics default driver sqlite, got %s", cfg.Databases.Analytics.Driver)
	}
	if cfg.Databases.Analytics.DSN != "file:./data/analytics.db" {
		t.Errorf("expected analytics default dsn file:./data/analytics.db, got %s", cfg.Databases.Analytics.DSN)
	}
	if cfg.Databases.Main.ReadOnly != false {
		t.Errorf("expected default read_only false, got true")
	}
	if cfg.Databases.Main.MaxIdleConns != 2 {
		t.Errorf("expected default max idle conns 2, got %d", cfg.Databases.Main.MaxIdleConns)
	}
	if cfg.Databases.Main.ConnMaxLifetime != 0 {
		t.Errorf("expected default lifetime 0, got %v", cfg.Databases.Main.ConnMaxLifetime)
	}
	if cfg.Databases.Main.ConnMaxIdleTime != 0 {
		t.Errorf("expected default idle time 0, got %v", cfg.Databases.Main.ConnMaxIdleTime)
	}
}
