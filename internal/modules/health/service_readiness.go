package health

import (
	"context"
	"errors"
	"time"
)

var errDatabaseDown = errors.New("Database check failed")

func (m *Module) getReadiness() (readinessResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if _, err := m.db.QueryContext(ctx, "SELECT 1"); err != nil {
		return readinessResponse{Status: readinessStatusDown}, errDatabaseDown
	}

	return readinessResponse{Status: readinessStatusUp}, nil
}
