package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
)

type ApplicationConfig struct {
	Environment          string        `validate:"oneof=development staging production"`
	Port                 int           `validate:"min=1,max=65535"`
	LogLevel             zerolog.Level `validate:"min=-1,max=5"`
	DebugMode            bool          `validate:"-"`
	CorsEnabled          bool          `validate:"-"`
	CorsAllowedOrigins   []string      `validate:"-"`
	CorsAllowedMethods   []string      `validate:"-"`
	CorsAllowedHeaders   []string      `validate:"-"`
	CompressionEnabled   bool          `validate:"-"`
	CompressionEncodings []string      `validate:"min=1,dive,oneof=zstd br gzip"`
	CompressionMinBytes  int           `validate:"min=0"`
	SwaggerEnabled       bool          `validate:"-"`
	SwaggerBasePath      string        `validate:"-"`
	SwaggerServerURL     string        `validate:"-"`
}

func parseApplicationConfig(values map[string]string) (ApplicationConfig, error) {
	config := ApplicationConfig{
		Environment:          "development",
		Port:                 8080,
		LogLevel:             zerolog.InfoLevel,
		DebugMode:            false,
		CorsEnabled:          true,
		CorsAllowedOrigins:   []string{"*"},
		CorsAllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		CorsAllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization"},
		CompressionEnabled:   false,
		CompressionEncodings: []string{"zstd", "br", "gzip"},
		CompressionMinBytes:  1024,
		SwaggerEnabled:       false,
		SwaggerBasePath:      "/swagger",
	}

	if raw, ok := values["APP_ENV"]; ok {
		config.Environment = strings.TrimSpace(raw)
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

	if raw, ok := values["APP_CORS_ENABLED"]; ok {
		corsEnabled, err := strconv.ParseBool(raw)
		if err != nil {
			return ApplicationConfig{}, fmt.Errorf("invalid APP_CORS_ENABLED %q: %w", raw, err)
		}
		config.CorsEnabled = corsEnabled
	}

	if raw, ok := values["APP_CORS_ALLOWED_ORIGINS"]; ok {
		config.CorsAllowedOrigins = splitList(raw)
	}

	if raw, ok := values["APP_CORS_ALLOWED_METHODS"]; ok {
		config.CorsAllowedMethods = splitList(raw)
	}

	if raw, ok := values["APP_CORS_ALLOWED_HEADERS"]; ok {
		config.CorsAllowedHeaders = splitList(raw)
	}

	if raw, ok := values["APP_COMPRESSION_ENABLED"]; ok {
		compressionEnabled, err := strconv.ParseBool(raw)
		if err != nil {
			return ApplicationConfig{}, fmt.Errorf("invalid APP_COMPRESSION_ENABLED %q: %w", raw, err)
		}
		config.CompressionEnabled = compressionEnabled
	}

	if raw, ok := values["APP_COMPRESSION_ENCODINGS"]; ok {
		encodings := splitList(raw)
		for i, enc := range encodings {
			encodings[i] = strings.ToLower(enc)
		}
		config.CompressionEncodings = encodings
	}

	if raw, ok := values["APP_COMPRESSION_MIN_BYTES"]; ok {
		minBytes, err := strconv.Atoi(raw)
		if err != nil {
			return ApplicationConfig{}, fmt.Errorf("invalid APP_COMPRESSION_MIN_BYTES %q: %w", raw, err)
		}
		config.CompressionMinBytes = minBytes
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

	if raw, ok := values["APP_SWAGGER_SERVER_URL"]; ok {
		config.SwaggerServerURL = strings.TrimSpace(raw)
	}

	if err := validator.New().Struct(config); err != nil {
		return ApplicationConfig{}, fmt.Errorf("validate application config: %w", err)
	}
	return config, nil
}
