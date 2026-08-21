package config

import (
	"fmt"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

type ApplicationConfig struct {
	Port      int           `validate:"min=1,max=65535"`
	LogLevel  zerolog.Level `validate:"min=-1,max=5"`
	DebugMode bool          `validate:"-"`
}

func parseApplicationConfig(values map[string]string) (ApplicationConfig, error) {
	config := ApplicationConfig{
		Port:      8080,
		LogLevel:  zerolog.InfoLevel,
		DebugMode: false,
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

	if err := validator.New().Struct(config); err != nil {
		return ApplicationConfig{}, fmt.Errorf("validate application config: %w", err)
	}
	return config, nil
}
