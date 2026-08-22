package probe

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) getReadiness(ctx *gin.Context) {
	response := c.service.GetReadiness(ctx.Request.Context())
	if response.DatabaseMain != ReadinessStatusOK || response.DatabaseAnalytics != ReadinessStatusOK {
		ctx.JSON(http.StatusServiceUnavailable, response)
		return
	}

	ctx.JSON(http.StatusOK, response)
}
