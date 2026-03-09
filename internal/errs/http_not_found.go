package errs

import "net/http"

func NewNotFound(msg string) *HTTPError {
	return &HTTPError{
		StatusCode: http.StatusNotFound,
		Code:       "NOT_FOUND",
		Message:    msg,
	}
}

// --- Only for Swagger ---

type HTTPNotFoundError struct {
	Code    string `json:"code"    example:"NOT_FOUND"`
	Message string `json:"message" example:"Route GET / not found"`
}
