package errs

import "net/http"

func NewServiceUnavailable() *HTTPError {
	return &HTTPError{
		StatusCode: http.StatusServiceUnavailable,
		Code:       "SERVICE_UNAVAILABLE",
		Message:    "Service unavailable",
	}
}

// --- Only for Swagger ---

type HTTPServiceUnavailableError struct {
	Code    string `json:"code"    example:"SERVICE_UNAVAILABLE"`
	Message string `json:"message" example:"Service unavailable"`
}
