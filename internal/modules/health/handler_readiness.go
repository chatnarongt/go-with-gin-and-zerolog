package health

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Check service readiness
// @Description Check connection to database and other services
// @Tags Health
// @Produce json
// @Success 200 {object} readinessResponseUp
// @Failure 503 {object} readinessResponseDown
// @Router /api/v1/health/readiness [get]
func (m *Module) readinessHandler(c *gin.Context) {
	result, err := m.getReadiness()

	if err != nil {
		if errors.Is(err, errDatabaseDown) {
			c.JSON(http.StatusServiceUnavailable, result)
			return
		}

		c.JSON(
			http.StatusInternalServerError,
			readinessResponse{
				Status: readinessStatusDown,
			},
		)
		return
	}

	c.JSON(http.StatusOK, result)
}
