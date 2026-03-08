package errs

type HTTPError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
}

func (e *HTTPError) Error() string {
	return e.Message
}

// --- Only for Swagger ---

type HTTPBadRequestError struct {
	Code    string `json:"code"    example:"BAD_REQUEST"`
	Message string `json:"message" example:"Bad request"`
}

type HTTPUnauthorizedError struct {
	Code    string `json:"code"    example:"UNAUTHORIZED"`
	Message string `json:"message" example:"Unauthorized"`
}

type HTTPNotFoundError struct {
	Code    string `json:"code"    example:"NOT_FOUND"`
	Message string `json:"message" example:"Route GET / not found"`
}

type HTTPInternalServerError struct {
	Code    string `json:"code"    example:"INTERNAL_SERVER_ERROR"`
	Message string `json:"message" example:"Internal server error"`
}

type HTTPServiceUnavailableError struct {
	Code    string `json:"code"    example:"SERVICE_UNAVAILABLE"`
	Message string `json:"message" example:"Service unavailable"`
}

var (
	_ = HTTPBadRequestError{}
	_ = HTTPUnauthorizedError{}
	_ = HTTPNotFoundError{}
	_ = HTTPInternalServerError{}
	_ = HTTPServiceUnavailableError{}
)
