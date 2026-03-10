package config

import (
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

func getEnvAsString(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		} else {
			log.Warn().Str("key", key).Str("value", value).Msg("Failed to parse integer, using fallback")
		}
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		} else {
			log.Warn().Str("key", key).Str("value", value).Msg("Failed to parse boolean, using fallback")
		}
	}
	return fallback
}
