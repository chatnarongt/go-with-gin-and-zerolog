package middleware

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/errs"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		req := c.Request

		var bodyBytes []byte
		if req.Body != nil {
			bodyBytes, _ = io.ReadAll(io.LimitReader(req.Body, 5<<20)) // 5 MB Cap
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		if httpError, ok := err.(*errs.HTTPError); ok {
			c.JSON(httpError.StatusCode, httpError)
			return
		}

		internalServerError := errs.NewInternalServerError()
		c.JSON(internalServerError.StatusCode, internalServerError)

		event := log.Error().Str("method", req.Method).Str("path", req.URL.Path)
		if len(req.URL.RawQuery) > 0 {
			event.Str("query", req.URL.RawQuery)
		}
		if len(bodyBytes) > 0 {
			if json.Valid(bodyBytes) {
				event.RawJSON("body", bodyBytes)
			} else {
				event.Str("body", string(bodyBytes))
			}
		}
		event.Msg(err.Error())
	}
}
