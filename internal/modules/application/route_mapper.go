package application

import "github.com/gin-gonic/gin"

type RouteMapper interface {
	MapRoutes(router *gin.Engine)
}

func (m *Module) MapRoutes(routeMappers ...RouteMapper) {
	for _, mapper := range routeMappers {
		mapper.MapRoutes(m.engine)
	}
}

type APIRouteMapper interface {
	MapAPIRoutes(router *gin.RouterGroup)
}

func (m *Module) MapAPIRoutes(routeMappers ...APIRouteMapper) {
	for _, mapper := range routeMappers {
		mapper.MapAPIRoutes(m.Router)
	}
}
