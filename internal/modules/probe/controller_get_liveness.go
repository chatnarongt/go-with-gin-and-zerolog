package probe

import "github.com/gin-gonic/gin"

func (c *Controller) getLiveness(ctx *gin.Context) {
	ctx.JSON(200, c.service.GetLiveness())
}
