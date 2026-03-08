package application

import (
	"fmt"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/errs"
	"github.com/gin-gonic/gin"
)

func notFound(c *gin.Context) {
	method := c.Request.Method

	path := c.Request.URL.Path

	msg := fmt.Sprintf("Route %s %s not found", method, path)

	c.Error(errs.NewNotFound(msg))
}
