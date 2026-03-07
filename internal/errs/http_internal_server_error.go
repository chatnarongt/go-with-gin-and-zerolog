package errs

import "net/http"

func NewInternalServerError() *HTTPError {
	return &HTTPError{
		StatusCode: http.StatusInternalServerError,
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "Internal server error",
	}
}
