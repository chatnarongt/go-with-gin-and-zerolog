package health

import "github.com/gin-gonic/gin"

func (m *Module) MapRoutes(router *gin.RouterGroup) {
	api := router.Group("/v1/health")
	api.GET("/liveness", m.livenessHandler)
	api.GET("/readiness", m.readinessHandler)
}
