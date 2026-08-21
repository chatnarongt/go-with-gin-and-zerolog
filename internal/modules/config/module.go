package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
)

type ModuleOptions struct {
	// EnvFiles is a list of paths to .env files to load.
	//
	// The files will be loaded in order, and later files will override earlier ones.
	EnvFiles []string
}

type Module struct {
	envFiles []string
	logger   *zerolog.Logger
}

var _ internal.Module = (*Module)(nil)

func NewModule(options ModuleOptions) *Module {
	return &Module{envFiles: options.EnvFiles}
}

func (m *Module) Register(i do.Injector, _ *gin.Engine) error {
	values, err := loadEnvFiles(m.envFiles)
	if err != nil {
		return err
	}

	for _, value := range os.Environ() {
		key, _, ok := strings.Cut(value, "=")
		if ok {
			values[key] = os.Getenv(key)
		}
	}

	application, err := parseApplicationConfig(values)
	if err != nil {
		return err
	}

	database, err := parseDatabaseConfig(values)
	if err != nil {
		return err
	}

	do.ProvideValue(i, &Config{Values: values, Application: application, Database: database})
	m.logger = do.MustInvoke[*zerolog.Logger](i)
	return nil
}

func (m *Module) OnModuleInit() error {
	m.logger.Info().Msg("Config module initialized")
	return nil
}

func loadEnvFiles(paths []string) (map[string]string, error) {
	values := make(map[string]string)
	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open env file %q: %w", path, err)
		}

		scanner := bufio.NewScanner(file)
		line := 0
		for scanner.Scan() {
			line++
			text := strings.TrimSpace(scanner.Text())
			if text == "" || strings.HasPrefix(text, "#") {
				continue
			}
			key, value, ok := strings.Cut(text, "=")
			if !ok || strings.TrimSpace(key) == "" {
				file.Close()
				return nil, fmt.Errorf("invalid env entry in %q at line %d", path, line)
			}
			values[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), "\"'")
		}
		if err := scanner.Err(); err != nil {
			file.Close()
			return nil, fmt.Errorf("read env file %q: %w", path, err)
		}
		if err := file.Close(); err != nil {
			return nil, fmt.Errorf("close env file %q: %w", path, err)
		}
	}
	return values, nil
}
