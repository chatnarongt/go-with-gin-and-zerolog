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
	if cfg.Databases.Main.Required != true {
		t.Errorf("expected default required true, got false")
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

	if cfg.Databases.Logging.Driver != "mongodb" {
		t.Errorf("expected logging default driver mongodb, got %s", cfg.Databases.Logging.Driver)
	}
	if cfg.Databases.Logging.DSN != "mongodb://localhost:27017/logs" {
		t.Errorf("expected logging default dsn mongodb://localhost:27017/logs, got %s", cfg.Databases.Logging.DSN)
	}
	if cfg.Databases.Logging.Required != false {
		t.Errorf("expected logging default required false, got true")
	}
}

func TestConfigModule_OptionalDatabaseEmptyDSNAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Setenv("DB_LOGGING_DSN", "")

	injector := do.New()
	nopLogger := zerolog.Nop()
	do.ProvideValue(injector, &nopLogger)

	cfgModule := config.NewModule(config.ModuleOptions{})
	if err := cfgModule.Register(injector, nil); err != nil {
		t.Fatalf("expected empty DSN for optional DB to pass, got err: %v", err)
	}

	cfg := do.MustInvoke[*config.Config](injector)
	if cfg.Databases.Logging.DSN != "" {
		t.Errorf("expected empty logging DSN, got %s", cfg.Databases.Logging.DSN)
	}
}

func TestConfigModule_CompressionDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	injector := do.New()
	nopLogger := zerolog.Nop()
	do.ProvideValue(injector, &nopLogger)

	cfgModule := config.NewModule(config.ModuleOptions{})
	if err := cfgModule.Register(injector, nil); err != nil {
		t.Fatalf("register config module: %v", err)
	}

	cfg := do.MustInvoke[*config.Config](injector)
	if cfg.Application.CompressionEnabled != false {
		t.Errorf("expected default CompressionEnabled false, got %v", cfg.Application.CompressionEnabled)
	}
	if len(cfg.Application.CompressionEncodings) != 3 ||
		cfg.Application.CompressionEncodings[0] != "zstd" ||
		cfg.Application.CompressionEncodings[1] != "br" ||
		cfg.Application.CompressionEncodings[2] != "gzip" {
		t.Errorf("expected default CompressionEncodings [zstd br gzip], got %v", cfg.Application.CompressionEncodings)
	}
	if cfg.Application.CompressionMinBytes != 1024 {
		t.Errorf("expected default CompressionMinBytes 1024, got %d", cfg.Application.CompressionMinBytes)
	}
}

func TestConfigModule_CompressionCustomValid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Setenv("APP_COMPRESSION_ENABLED", "true")
	t.Setenv("APP_COMPRESSION_ENCODINGS", "gzip, BR")
	t.Setenv("APP_COMPRESSION_MIN_BYTES", "2048")

	injector := do.New()
	nopLogger := zerolog.Nop()
	do.ProvideValue(injector, &nopLogger)

	cfgModule := config.NewModule(config.ModuleOptions{})
	if err := cfgModule.Register(injector, nil); err != nil {
		t.Fatalf("register config module: %v", err)
	}

	cfg := do.MustInvoke[*config.Config](injector)
	if !cfg.Application.CompressionEnabled {
		t.Errorf("expected CompressionEnabled true, got false")
	}
	if len(cfg.Application.CompressionEncodings) != 2 ||
		cfg.Application.CompressionEncodings[0] != "gzip" ||
		cfg.Application.CompressionEncodings[1] != "br" {
		t.Errorf("expected CompressionEncodings [gzip br], got %v", cfg.Application.CompressionEncodings)
	}
	if cfg.Application.CompressionMinBytes != 2048 {
		t.Errorf("expected CompressionMinBytes 2048, got %d", cfg.Application.CompressionMinBytes)
	}
}

func TestConfigModule_CompressionInvalidValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("invalid encoding", func(t *testing.T) {
		t.Setenv("APP_COMPRESSION_ENCODINGS", "deflate,gzip")

		injector := do.New()
		nopLogger := zerolog.Nop()
		do.ProvideValue(injector, &nopLogger)

		cfgModule := config.NewModule(config.ModuleOptions{})
		if err := cfgModule.Register(injector, nil); err == nil {
			t.Errorf("expected error for invalid encoding 'deflate', got nil")
		}
	})

	t.Run("negative min bytes", func(t *testing.T) {
		t.Setenv("APP_COMPRESSION_MIN_BYTES", "-1")

		injector := do.New()
		nopLogger := zerolog.Nop()
		do.ProvideValue(injector, &nopLogger)

		cfgModule := config.NewModule(config.ModuleOptions{})
		if err := cfgModule.Register(injector, nil); err == nil {
			t.Errorf("expected error for negative min bytes, got nil")
		}
	})
}
