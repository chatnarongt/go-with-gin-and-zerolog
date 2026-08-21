package probe

import "github.com/gin-gonic/gin"

func (c *Controller) getReadiness(ctx *gin.Context) {
	response := c.service.GetReadiness(ctx.Request.Context())
	status := 200
	if response.Database != ReadinessStatusOK {
		status = 503
	}
	ctx.JSON(status, response)
}
