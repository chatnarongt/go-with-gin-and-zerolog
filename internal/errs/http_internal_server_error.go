package errs

import "net/http"

func NewInternalServerError() *HTTPError {
	return &HTTPError{
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "Internal server error",
	}
}

// --- Only for Swagger ---

type HTTPInternalServerError struct {
	Code    string `json:"code"    example:"INTERNAL_SERVER_ERROR"`
	Message string `json:"message" example:"Internal server error"`
}
