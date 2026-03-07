package health

import "github.com/gin-gonic/gin"

func (m *Module) setupController(router *gin.RouterGroup) {
	api := router.Group("/v1")
	api.GET("/liveness", m.liveness)
	api.GET("/readiness", m.readiness)
}
