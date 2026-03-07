package errs

import "net/http"

func NewBadRequest(msg string) *HTTPError {
	return &HTTPError{
		StatusCode: http.StatusBadRequest,
		Code:       "BAD_REQUEST",
		Message:    msg,
	}
}
