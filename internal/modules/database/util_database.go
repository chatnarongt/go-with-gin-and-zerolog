package database

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func normalizeDriver(driver string) string {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "postgres", "pgx":
		return "pgx"
	case "sqlserver", "mssql":
		return "sqlserver"
	case "sqlite":
		return "sqlite"
	case "mongodb", "mongo":
		return "mongodb"
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
	case "mongodb":
		if strings.Contains(strings.ToLower(dsn), "readpreference") {
			return dsn
		}
		separator := "?"
		if strings.Contains(dsn, "?") {
			separator = "&"
		}
		return dsn + separator + "readPreference=secondary"
	default:
		return dsn
	}
}

func parseMongoDatabaseName(dsn string) (string, error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return "", fmt.Errorf("invalid mongodb DSN %q: %w", dsn, err)
	}
	dbName := strings.TrimPrefix(parsed.Path, "/")
	if dbName == "" {
		return "", fmt.Errorf("mongodb DSN %q requires a database name in the path (e.g. mongodb://host:port/dbname)", dsn)
	}
	return dbName, nil
}

func ensureDatabaseDirectory(dsn string) error {
	if !strings.HasPrefix(dsn, "file:") {
		return nil
	}

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
