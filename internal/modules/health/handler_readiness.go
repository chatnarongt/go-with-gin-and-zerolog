package health

import (
	"errors"
	"net/http"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/errs"
	"github.com/gin-gonic/gin"
)

// @Summary Check service readiness
// @Description Check connection to database and other services
// @Tags Health
// @Produce json
// @Success 200 {object} readinessResponseUp
// @Failure 500 {object} errs.HTTPError
// @Failure 503 {object} readinessResponseDown
// @Router /v1/readinessHandler [get]
func (m *Module) readinessHandler(c *gin.Context) {
	result, err := m.getReadiness()

	if err != nil {
		if errors.Is(err, errDatabaseDown) {
			c.JSON(http.StatusServiceUnavailable, result)
			return
		}

		c.Error(errs.NewInternalServerError())
		return
	}

	c.JSON(http.StatusOK, result)
}
