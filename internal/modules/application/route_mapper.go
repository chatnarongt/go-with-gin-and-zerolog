package application

import "github.com/gin-gonic/gin"

type RouteMapper interface {
	MapRoutes(router *gin.RouterGroup)
}

func (m *Module) MapRoutes(routeMappers ...RouteMapper) {
	for _, mapper := range routeMappers {
		mapper.MapRoutes(m.Router)
	}
}
