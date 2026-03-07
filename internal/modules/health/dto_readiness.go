package health

type readinessStatus string

const (
	readinessStatusUp   readinessStatus = "UP"
	readinessStatusDown readinessStatus = "DOWN"
)

type readinessResponse struct {
	Status readinessStatus `json:"status"`
}

// --- Only for Swagger ---

type readinessResponseUp struct {
	Status readinessStatus `json:"status" example:"UP"`
}
type readinessResponseDown struct {
	Status readinessStatus `json:"status" example:"DOWN"`
}

var _ = readinessResponseUp{}
var _ = readinessResponseDown{}
