package swagger

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	"github.com/swaggest/swgui"
	swaggerui "github.com/swaggest/swgui/v5"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/modules/config"
)

type ModuleOptions struct {
	Spec        []byte
	Title       string
	Version     string
	Description string
}

type Module struct {
	options ModuleOptions
	logger  *zerolog.Logger
	enabled bool
}

func NewModule(options ModuleOptions) *Module {
	return &Module{options: options}
}

var _ internal.Module = (*Module)(nil)

func (m *Module) Register(i do.Injector, router *gin.Engine) error {
	applicationConfig := do.MustInvoke[*config.Config](i).Application
	m.logger = do.MustInvoke[*zerolog.Logger](i)
	if !applicationConfig.SwaggerEnabled {
		return nil
	}
	m.enabled = true
	if m.options.Title == "" {
		return fmt.Errorf("swagger title is required")
	}
	if m.options.Version == "" {
		return fmt.Errorf("swagger version is required")
	}

	spec, err := loadSpec(m.options.Spec)
	if err != nil {
		return err
	}
	spec.Info.Title = m.options.Title
	spec.Info.Version = m.options.Version
	spec.Info.Description = m.options.Description

	specJSON, err := json.Marshal(spec)
	if err != nil {
		return fmt.Errorf("marshal OpenAPI spec: %w", err)
	}

	router.GET("/openapi.json", func(ctx *gin.Context) {
		ctx.Data(http.StatusOK, "application/json; charset=utf-8", specJSON)
	})

	basePath := applicationConfig.SwaggerBasePath
	ui := swaggerui.NewHandlerWithConfig(swgui.Config{
		Title:       m.options.Title,
		SwaggerJSON: "/openapi.json",
		BasePath:    basePath,
	})
	router.GET(basePath, gin.WrapH(ui))
	router.GET(basePath+"/*any", gin.WrapH(ui))
	return nil
}

func (m *Module) OnModuleInit() error {
	if !m.enabled {
		return nil
	}
	m.logger.Info().Msg("Swagger module initialized")
	return nil
}

func loadSpec(data []byte) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false
	spec, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("load OpenAPI spec: %w", err)
	}
	if err := spec.Validate(context.Background()); err != nil {
		return nil, fmt.Errorf("validate OpenAPI spec: %w", err)
	}
	return spec, nil
}
