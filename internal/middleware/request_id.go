package middleware

import (
	"context"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/errs"
	"github.com/gin-gonic/gin"
)

const requestIDHeader = "X-Request-ID"

type requestIDContextKey struct{}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			var err error
			requestID, err = internal.NewID()
			if err != nil {
				response := errs.InternalServerError().Response()
				c.AbortWithStatusJSON(response.Status, response)
				return
			}
		}

		c.Header(requestIDHeader, requestID)
		c.Set(requestIDHeader, requestID)
		c.Request = c.Request.WithContext(
			context.WithValue(c.Request.Context(), requestIDContextKey{}, requestID),
		)
		c.Next()
	}
}

func RequestIDFromContext(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDContextKey{}).(string)
	return requestID, ok
}
