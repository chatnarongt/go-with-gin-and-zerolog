package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

type ApplicationConfig struct {
	Port            int           `validate:"min=1,max=65535"`
	LogLevel        zerolog.Level `validate:"min=-1,max=5"`
	DebugMode       bool          `validate:"-"`
	SwaggerEnabled  bool          `validate:"-"`
	SwaggerBasePath string        `validate:"-"`
}

func parseApplicationConfig(values map[string]string) (ApplicationConfig, error) {
	config := ApplicationConfig{
		Port:            8080,
		LogLevel:        zerolog.InfoLevel,
		DebugMode:       false,
		SwaggerEnabled:  false,
		SwaggerBasePath: "/swagger",
	}

	if raw, ok := values["APP_PORT"]; ok {
		port, err := strconv.Atoi(raw)
		if err != nil {
			return ApplicationConfig{}, fmt.Errorf("invalid APP_PORT %q: %w", raw, err)
		}
		config.Port = port
	}

	if raw, ok := values["APP_LOG_LEVEL"]; ok {
		level, err := strconv.Atoi(raw)
		if err != nil {
			return ApplicationConfig{}, fmt.Errorf("invalid APP_LOG_LEVEL %q: %w", raw, err)
		}
		config.LogLevel = zerolog.Level(level)
	}

	if raw, ok := values["APP_DEBUG_MODE"]; ok {
		debugMode, err := strconv.ParseBool(raw)
		if err != nil {
			return ApplicationConfig{}, fmt.Errorf("invalid APP_DEBUG_MODE %q: %w", raw, err)
		}
		config.DebugMode = debugMode
	}

	if raw, ok := values["APP_SWAGGER_ENABLED"]; ok {
		swaggerEnabled, err := strconv.ParseBool(raw)
		if err != nil {
			return ApplicationConfig{}, fmt.Errorf("invalid APP_SWAGGER_ENABLED %q: %w", raw, err)
		}
		config.SwaggerEnabled = swaggerEnabled
	}

	if raw, ok := values["APP_SWAGGER_BASE_PATH"]; ok {
		basePath := strings.TrimRight(strings.TrimSpace(raw), "/")
		if basePath == "" || !strings.HasPrefix(basePath, "/") || strings.ContainsAny(basePath, "?#") {
			return ApplicationConfig{}, fmt.Errorf("invalid APP_SWAGGER_BASE_PATH %q: must be an absolute path", raw)
		}
		config.SwaggerBasePath = basePath
	}

	if err := validator.New().Struct(config); err != nil {
		return ApplicationConfig{}, fmt.Errorf("validate application config: %w", err)
	}
	return config, nil
}
