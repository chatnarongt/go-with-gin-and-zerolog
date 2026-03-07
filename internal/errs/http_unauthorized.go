package errs

import "net/http"

func NewUnauthorized(msg string) *HTTPError {
	return &HTTPError{
		StatusCode: http.StatusUnauthorized,
		Code:       "UNAUTHORIZED",
		Message:    msg,
	}
}
