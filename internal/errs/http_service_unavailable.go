package errs

import "net/http"

func NewServiceUnavailable() *HTTPError {
	return &HTTPError{
		StatusCode: http.StatusServiceUnavailable,
		Code:       "SERVICE_UNAVAILABLE",
		Message:    "Service unavailable",
	}
}
