package swagger

import (
	"net/http"

	_ "github.com/chatnarongt/go-with-gin-and-zerolog/docs"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (m *Module) MapRoutes(router *gin.Engine) {
	appConfig := m.config.LoadAppConfig()

	if appConfig.EnableSwagger {
		config := &ginSwagger.Config{
			URL:                      "/swagger/doc.json",
			DeepLinking:              true,
			DocExpansion:             "list",
			DefaultModelsExpandDepth: 0,
		}
		router.GET("/swagger", func(c *gin.Context) {
			c.Redirect(http.StatusTemporaryRedirect, "/swagger/index.html")
		})
		router.GET("/swagger/*any", ginSwagger.CustomWrapHandler(config, swaggerFiles.Handler))
	}
}
