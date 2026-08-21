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
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
)

type Detail struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type Response struct {
	Status  int      `json:"status"`
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Errors  []Detail `json:"errors,omitempty"`
}

type Error struct {
	response Response
	cause    error
}

func New(status int, code, message string, details ...Detail) *Error {
	return &Error{
		response: Response{
			Status:  status,
			Code:    code,
			Message: message,
			Errors:  details,
		},
	}
}

func Wrap(cause error, status int, code, message string, details ...Detail) *Error {
	httpError := New(status, code, message, details...)
	httpError.cause = cause
	return httpError
}

func BadRequest(message string, details ...Detail) *Error {
	return New(http.StatusBadRequest, CodeBadRequest, message, details...)
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
	response.Errors = append([]Detail(nil), e.response.Errors...)
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
		return BadRequest("Invalid request body.", Detail{
			Path:    "",
			Message: "Request body is required.",
		}).Response()
	}

	var syntaxError *json.SyntaxError
	if errors.As(err, &syntaxError) {
		return BadRequest("Invalid request body.", Detail{
			Path:    "",
			Message: "Malformed JSON.",
		}).Response()
	}

	var typeError *json.UnmarshalTypeError
	if errors.As(err, &typeError) {
		return BadRequest("Invalid request body.", Detail{
			Path:    lowerCamel(typeError.Field),
			Message: "Invalid value.",
		}).Response()
	}

	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		return BadRequest("Invalid request body.", Detail{
			Path:    "",
			Message: "Request body is too large.",
		}).Response()
	}

	return InternalServerError().Response()
}

func validationDetails(validationErrors validator.ValidationErrors) []Detail {
	details := make([]Detail, 0, len(validationErrors))
	for _, validationError := range validationErrors {
		details = append(details, Detail{
			Path:    fieldPath(validationError.Namespace()),
			Message: validationMessage(validationError),
		})
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
