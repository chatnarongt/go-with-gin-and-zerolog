package health

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

type Module struct {
	db *sql.DB
}

func NewModule(db *sql.DB) *Module {
	return &Module{db: db}
}

func (m *Module) Register(router *gin.RouterGroup) {
	m.setupController(router)
}
