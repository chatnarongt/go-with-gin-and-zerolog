package config

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
)

const defaultDatabaseDSN = "file:./data/app.db"

type DatabaseConfig struct {
	DSN             string        `validate:"required"`
	MaxOpenConns    int           `validate:"min=1"`
	MaxIdleConns    int           `validate:"min=0"`
	ConnMaxLifetime time.Duration `validate:"min=0"`
	ConnMaxIdleTime time.Duration `validate:"min=0"`
}

func parseDatabaseConfig(values map[string]string) (DatabaseConfig, error) {
	config := DatabaseConfig{
		DSN:             defaultDatabaseDSN,
		MaxOpenConns:    0,
		MaxIdleConns:    2,
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 0,
	}

	if raw, ok := values["DATABASE_DSN"]; ok {
		config.DSN = raw
	}

	if raw, ok := values["DATABASE_MAX_OPEN_CONNS"]; ok {
		maxOpenConns, err := strconv.Atoi(raw)
		if err != nil {
			return DatabaseConfig{}, fmt.Errorf("invalid DATABASE_MAX_OPEN_CONNS %q: %w", raw, err)
		}
		config.MaxOpenConns = maxOpenConns
	}

	if raw, ok := values["DATABASE_MAX_IDLE_CONNS"]; ok {
		maxIdleConns, err := strconv.Atoi(raw)
		if err != nil {
			return DatabaseConfig{}, fmt.Errorf("invalid DATABASE_MAX_IDLE_CONNS %q: %w", raw, err)
		}
		config.MaxIdleConns = maxIdleConns
	}

	for key, target := range map[string]*time.Duration{
		"DATABASE_MAX_LIFETIME":  &config.ConnMaxLifetime,
		"DATABASE_MAX_IDLE_TIME": &config.ConnMaxIdleTime,
	} {
		if raw, ok := values[key]; ok {
			duration, err := time.ParseDuration(raw)
			if err != nil {
				return DatabaseConfig{}, fmt.Errorf("invalid %s %q: %w", key, raw, err)
			}
			*target = duration
		}
	}

	if err := validator.New().Struct(config); err != nil {
		return DatabaseConfig{}, fmt.Errorf("validate database config: %w", err)
	}
	if config.MaxIdleConns > config.MaxOpenConns {
		return DatabaseConfig{}, fmt.Errorf("validate database config: max idle connections %d exceeds max open connections %d", config.MaxIdleConns, config.MaxOpenConns)
	}

	return config, nil
}
