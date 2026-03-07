package errs

import "net/http"

func NewNotFound(msg string) *HTTPError {
	return &HTTPError{
		StatusCode: http.StatusNotFound,
		Code:       "NOT_FOUND",
		Message:    msg,
	}
}
