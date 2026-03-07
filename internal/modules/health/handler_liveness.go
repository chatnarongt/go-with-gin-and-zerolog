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
// @Router /v1/liveness [get]
func (m *Module) liveness(c *gin.Context) {
	result := m.getLiveness()
	c.String(http.StatusOK, string(result))
}
