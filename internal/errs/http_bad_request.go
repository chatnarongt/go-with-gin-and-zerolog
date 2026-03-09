package errs

import "net/http"

func NewBadRequest(msg string) *HTTPError {
	return &HTTPError{
		StatusCode: http.StatusBadRequest,
		Code:       "BAD_REQUEST",
		Message:    msg,
	}
}

// --- Only for Swagger ---

type HTTPBadRequestError struct {
	Code    string `json:"code"    example:"BAD_REQUEST"`
	Message string `json:"message" example:"Bad request"`
}
