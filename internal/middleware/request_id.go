package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

const requestIDHeader = "X-Request-ID"

type requestIDContextKey struct{}

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID, err := newRequestID()
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
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

func newRequestID() (string, error) {
	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}
	return hex.EncodeToString(id), nil
}
