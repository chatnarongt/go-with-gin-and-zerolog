package errs

import "net/http"

func NewUnauthorized(msg string) *HTTPError {
	return &HTTPError{
		StatusCode: http.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    msg,
	}
}

// --- Only for Swagger ---

type HTTPUnauthorizedError struct {
	Code    string `json:"code"    example:"UNAUTHORIZED"`
	Message string `json:"message" example:"Unauthorized"`
}
