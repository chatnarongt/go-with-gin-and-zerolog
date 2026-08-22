package errs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

const (
	CodeBadRequest          = "BAD_REQUEST"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeMethodNotAllowed    = "METHOD_NOT_ALLOWED"
	CodeNotImplemented      = "NOT_IMPLEMENTED"
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
	CodeBadGateway          = "BAD_GATEWAY"
	CodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
)

type Response struct {
	Status  int      `json:"status"`
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Errors  []string `json:"errors"`
}

type Error struct {
	response Response
	cause    error
}

func New(status int, code, message string, responseErrors ...string) *Error {
	return &Error{
		response: Response{
			Status:  status,
			Code:    code,
			Message: message,
			Errors:  nonEmpty(responseErrors),
		},
	}
}

func Wrap(cause error, status int, code, message string, responseErrors ...string) *Error {
	httpError := New(status, code, message, responseErrors...)
	httpError.cause = cause
	return httpError
}

func BadRequest(message string, responseErrors ...string) *Error {
	return New(http.StatusBadRequest, CodeBadRequest, message, responseErrors...)
}

func Unauthorized(message string, responseErrors ...string) *Error {
	return New(http.StatusUnauthorized, CodeUnauthorized, message, responseErrors...)
}

func Forbidden(message string, responseErrors ...string) *Error {
	return New(http.StatusForbidden, CodeForbidden, message, responseErrors...)
}

func NotFound(message string, responseErrors ...string) *Error {
	return New(http.StatusNotFound, CodeNotFound, message, responseErrors...)
}

func MethodNotAllowed(message string, responseErrors ...string) *Error {
	return New(http.StatusMethodNotAllowed, CodeMethodNotAllowed, message, responseErrors...)
}

func NotImplemented(message string, responseErrors ...string) *Error {
	return New(http.StatusNotImplemented, CodeNotImplemented, message, responseErrors...)
}

func BadGateway(message string, responseErrors ...string) *Error {
	return New(http.StatusBadGateway, CodeBadGateway, message, responseErrors...)
}

func ServiceUnavailable(message string, responseErrors ...string) *Error {
	return New(http.StatusServiceUnavailable, CodeServiceUnavailable, message, responseErrors...)
}

func InternalServerError() *Error {
	return New(http.StatusInternalServerError, CodeInternalServerError, "Internal server error.")
}

func (e *Error) Error() string {
	return e.response.Message
}

func (e *Error) Unwrap() error {
	return e.cause
}

func (e *Error) Response() Response {
	response := e.response
	response.Errors = append([]string(nil), e.response.Errors...)
	return response
}

func ResponseFor(err error) Response {
	if err == nil {
		return InternalServerError().Response()
	}

	var httpError *Error
	if errors.As(err, &httpError) && httpError != nil {
		return httpError.Response()
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		return BadRequest("Invalid request body.", validationDetails(validationErrors)...).Response()
	}

	if errors.Is(err, io.EOF) {
		return invalidBodyError("", "Request body is required.").Response()
	}

	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) {
		return invalidBodyError("", "Malformed JSON.").Response()
	}

	var typeError *json.UnmarshalTypeError
	if errors.As(err, &typeError) {
		return invalidBodyError(lowerCamel(typeError.Field), "Invalid value.").Response()
	}

	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		return invalidBodyError("", "Request body is too large.").Response()
	}

	return InternalServerError().Response()
}

func invalidBody(message string) *Error {
	return BadRequest("Invalid request body.", message)
}

func invalidBodyError(path, message string) *Error {
	if path == "" {
		return invalidBody(message)
	}
	return invalidBody(fmt.Sprintf("%s: %s", path, message))
}

func validationDetails(validationErrors validator.ValidationErrors) []string {
	details := make([]string, 0, len(validationErrors))
	for _, validationError := range validationErrors {
		details = append(details, fmt.Sprintf("%s: %s", fieldPath(validationError.Namespace()), validationMessage(validationError)))
	}
	return details
}

func fieldPath(namespace string) string {
	parts := strings.Split(namespace, ".")
	if len(parts) > 1 {
		parts = parts[1:]
	}
	for index, part := range parts {
		parts[index] = lowerCamel(part)
	}
	return strings.Join(parts, ".")
}

func lowerCamel(value string) string {
	if value == "" {
		return value
	}

	runes := []rune(value)
	end := 1
	for end < len(runes) && unicode.IsUpper(runes[end]) {
		if end+1 < len(runes) && unicode.IsLower(runes[end+1]) {
			break
		}
		end++
	}
	for index := 0; index < end; index++ {
		runes[index] = unicode.ToLower(runes[index])
	}
	return string(runes)
}

func validationMessage(validationError validator.FieldError) string {
	switch validationError.Tag() {
	case "required":
		return "This field is required."
	case "email":
		return "Must be a valid email address."
	case "url", "uri":
		return "Must be a valid URL."
	case "min":
		return fmt.Sprintf("Must be at least %s.", validationError.Param())
	case "max":
		return fmt.Sprintf("Must be at most %s.", validationError.Param())
	default:
		return "Invalid value."
	}
}

func nonEmpty(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	result := make([]string, 0, len(values))
	for _, value := range values {
		if value != "" {
			result = append(result, value)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
