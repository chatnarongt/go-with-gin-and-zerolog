package middleware

import (
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/errs"
	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recover() != nil {
				writeError(c, errs.InternalServerError().Response())
				return
			}

			if c.Writer.Written() || len(c.Errors) == 0 {
				return
			}

			writeError(c, errs.ResponseFor(c.Errors.Last().Err))
		}()

		c.Next()
	}
}

func writeError(c *gin.Context, response errs.Response) {
	c.AbortWithStatusJSON(response.Status, response)
}
