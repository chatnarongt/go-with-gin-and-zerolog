package errs

type HTTPError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *HTTPError) Error() string {
	return e.Message
}
