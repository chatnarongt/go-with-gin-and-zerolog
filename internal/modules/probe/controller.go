package probe

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do/v2"
)

type Controller struct {
	service *Service
}

func NewController(i do.Injector) (*Controller, error) {
	return &Controller{service: do.MustInvoke[*Service](i)}, nil
}

func (c *Controller) RegisterRoutes(router *gin.Engine) {
	router.GET("/probe/liveness", c.getLiveness)
	router.GET("/probe/readiness", c.getReadiness)
}
