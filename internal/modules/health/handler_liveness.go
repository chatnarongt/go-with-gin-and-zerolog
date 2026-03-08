package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary Check service health
// @Description Check service health return "OK"
// @Tags Health
// @Produce plain
// @Success 200 {object} livenessStatus
// @Router /api/v1/health/liveness [get]
func (m *Module) livenessHandler(c *gin.Context) {
	result := m.getLiveness()
	c.String(http.StatusOK, string(result))
}
