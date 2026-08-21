package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func AccessLog(log func() *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()

		c.Next()

		id, _ := RequestIDFromContext(c.Request.Context())

		timeTaken := time.Since(startedAt)

		event := log().Trace().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Str("requestId", id).
			Int("status", c.Writer.Status()).
			Str("timeTakenReadable", timeTaken.String()).
			Dur("timeTaken", timeTaken)

		if len(c.Errors) > 0 {
			event = event.Str("errors", c.Errors.String())
		}

		event.Msg("Access")
	}
}
