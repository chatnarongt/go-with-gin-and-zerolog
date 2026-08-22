package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

const (
	defaultMainDatabaseDriver      = "sqlite"
	defaultMainDatabaseDSN         = "file:./data/app.db"
	defaultAnalyticsDatabaseDriver = "sqlite"
	defaultAnalyticsDatabaseDSN    = "file:./data/analytics.db"
)

type DatabaseConnectionConfig struct {
	Driver          string        `validate:"required,oneof=sqlite postgres pgx sqlserver mssql"`
	DSN             string        `validate:"required"`
	ReadOnly        bool          `validate:"-"`
	MaxOpenConns    int           `validate:"min=0"`
	MaxIdleConns    int           `validate:"min=0"`
	ConnMaxLifetime time.Duration `validate:"min=0"`
	ConnMaxIdleTime time.Duration `validate:"min=0"`
}

type DatabasesConfig struct {
	Main      DatabaseConnectionConfig
	Analytics DatabaseConnectionConfig
}

func parseDatabaseConnection(prefix, defaultDriver, defaultDSN string, values map[string]string) (DatabaseConnectionConfig, error) {
	config := DatabaseConnectionConfig{
		Driver:          defaultDriver,
		DSN:             defaultDSN,
		ReadOnly:        false,
		MaxOpenConns:    0,
		MaxIdleConns:    2,
		ConnMaxLifetime: 0,
		ConnMaxIdleTime: 0,
	}

	if raw, ok := values[prefix+"_DRIVER"]; ok {
		config.Driver = strings.ToLower(strings.TrimSpace(raw))
	}

	if raw, ok := values[prefix+"_DSN"]; ok {
		config.DSN = strings.TrimSpace(raw)
	}

	if raw, ok := values[prefix+"_READ_ONLY"]; ok {
		readOnly, err := strconv.ParseBool(raw)
		if err != nil {
			return DatabaseConnectionConfig{}, fmt.Errorf("invalid %s_READ_ONLY %q: %w", prefix, raw, err)
		}
		config.ReadOnly = readOnly
	}

	if raw, ok := values[prefix+"_MAX_OPEN_CONNS"]; ok {
		maxOpenConns, err := strconv.Atoi(raw)
		if err != nil {
			return DatabaseConnectionConfig{}, fmt.Errorf("invalid %s_MAX_OPEN_CONNS %q: %w", prefix, raw, err)
		}
		config.MaxOpenConns = maxOpenConns
	}

	if raw, ok := values[prefix+"_MAX_IDLE_CONNS"]; ok {
		maxIdleConns, err := strconv.Atoi(raw)
		if err != nil {
			return DatabaseConnectionConfig{}, fmt.Errorf("invalid %s_MAX_IDLE_CONNS %q: %w", prefix, raw, err)
		}
		config.MaxIdleConns = maxIdleConns
	}

	for key, target := range map[string]*time.Duration{
		prefix + "_MAX_LIFETIME":  &config.ConnMaxLifetime,
		prefix + "_MAX_IDLE_TIME": &config.ConnMaxIdleTime,
	} {
		if raw, ok := values[key]; ok {
			duration, err := time.ParseDuration(raw)
			if err != nil {
				return DatabaseConnectionConfig{}, fmt.Errorf("invalid %s %q: %w", key, raw, err)
			}
			*target = duration
		}
	}

	if err := validator.New().Struct(config); err != nil {
		return DatabaseConnectionConfig{}, fmt.Errorf("validate %s config: %w", strings.ToLower(prefix), err)
	}
	if config.MaxOpenConns > 0 && config.MaxIdleConns > config.MaxOpenConns {
		return DatabaseConnectionConfig{}, fmt.Errorf("validate %s config: max idle connections %d exceeds max open connections %d", strings.ToLower(prefix), config.MaxIdleConns, config.MaxOpenConns)
	}

	return config, nil
}

func parseDatabasesConfig(values map[string]string) (DatabasesConfig, error) {
	mainConfig, err := parseDatabaseConnection("DB_MAIN", defaultMainDatabaseDriver, defaultMainDatabaseDSN, values)
	if err != nil {
		return DatabasesConfig{}, err
	}

	analyticsConfig, err := parseDatabaseConnection("DB_ANALYTICS", defaultAnalyticsDatabaseDriver, defaultAnalyticsDatabaseDSN, values)
	if err != nil {
		return DatabasesConfig{}, err
	}

	return DatabasesConfig{
		Main:      mainConfig,
		Analytics: analyticsConfig,
	}, nil
}
