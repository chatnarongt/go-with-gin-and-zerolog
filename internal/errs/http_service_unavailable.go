package errs

import "net/http"

func NewServiceUnavailable(msg string) *HTTPError {
	return &HTTPError{
		StatusCode: http.StatusServiceUnavailable,
		Code:       "SERVICE_UNAVAILABLE",
		Message:    msg,
	}
}
