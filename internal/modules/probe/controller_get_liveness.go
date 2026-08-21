package probe

import "github.com/gin-gonic/gin"

func (c *Controller) getLiveness(ctx *gin.Context) {
	result := c.service.GetLiveness()
	ctx.String(200, string(result))
}
