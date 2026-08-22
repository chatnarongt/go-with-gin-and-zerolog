package errs

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestResponseFor(t *testing.T) {
	validate := validator.New()
	validationErr := validate.Struct(struct {
		Email string `validate:"required,email"`
	}{}).(validator.ValidationErrors)
	type person struct {
		Email string `validate:"required,email"`
	}
	type request struct {
		Person person
	}
	nestedValidationErr := validate.Struct(request{}).(validator.ValidationErrors)

	tests := []struct {
		name     string
		err      error
		expected Response
	}{
		{
			name: "custom error",
			err:  BadRequest("Invalid request body.", "name: This field is required."),
			expected: Response{
				Status:  http.StatusBadRequest,
				Code:    CodeBadRequest,
				Message: "Invalid request body.",
				Errors:  []string{"name: This field is required."},
			},
		},
		{
			name: "validation error",
			err:  validationErr,
			expected: Response{
				Status:  http.StatusBadRequest,
				Code:    CodeBadRequest,
				Message: "Invalid request body.",
				Errors:  []string{"email: This field is required."},
			},
		},
		{
			name: "nested validation error",
			err:  nestedValidationErr,
			expected: Response{
				Status:  http.StatusBadRequest,
				Code:    CodeBadRequest,
				Message: "Invalid request body.",
				Errors:  []string{"person.email: This field is required."},
			},
		},
		{
			name: "syntax error",
			err:  &json.SyntaxError{},
			expected: Response{
				Status:  http.StatusBadRequest,
				Code:    CodeBadRequest,
				Message: "Invalid request body.",
				Errors:  []string{"Malformed JSON."},
			},
		},
		{
			name: "unknown error",
			err:  errors.New("database unavailable"),
			expected: Response{
				Status:  http.StatusInternalServerError,
				Code:    CodeInternalServerError,
				Message: "Internal server error.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResponseFor(tt.err)
			if got.Status != tt.expected.Status || got.Code != tt.expected.Code || got.Message != tt.expected.Message {
				t.Fatalf("ResponseFor(%v) = %#v, want %#v", tt.err, got, tt.expected)
			}
			if len(got.Errors) != len(tt.expected.Errors) {
				t.Fatalf("ResponseFor(%v) errors = %#v, want %#v", tt.err, got.Errors, tt.expected.Errors)
			}
			for index := range got.Errors {
				if got.Errors[index] != tt.expected.Errors[index] {
					t.Errorf("ResponseFor(%v) errors[%d] = %#v, want %#v", tt.err, index, got.Errors[index], tt.expected.Errors[index])
				}
			}
		})
	}
}

func TestHTTPErrorConstructors(t *testing.T) {
	tests := []struct {
		name       string
		newError   func(string, ...string) *Error
		statusCode int
		code       string
	}{
		{"unauthorized", Unauthorized, http.StatusUnauthorized, CodeUnauthorized},
		{"forbidden", Forbidden, http.StatusForbidden, CodeForbidden},
		{"not found", NotFound, http.StatusNotFound, CodeNotFound},
		{"method not allowed", MethodNotAllowed, http.StatusMethodNotAllowed, CodeMethodNotAllowed},
		{"not implemented", NotImplemented, http.StatusNotImplemented, CodeNotImplemented},
		{"bad gateway", BadGateway, http.StatusBadGateway, CodeBadGateway},
		{"service unavailable", ServiceUnavailable, http.StatusServiceUnavailable, CodeServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := tt.newError("Request failed.").Response()
			if response.Status != tt.statusCode || response.Code != tt.code {
				t.Fatalf("response = %#v, want status %d code %q", response, tt.statusCode, tt.code)
			}
		})
	}
}

func TestResponseJSONIncludesNullErrorsWhenEmpty(t *testing.T) {
	body, err := json.Marshal(InternalServerError().Response())
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"status":500,"code":"INTERNAL_SERVER_ERROR","message":"Internal server error.","errors":null}` {
		t.Fatalf("JSON = %s", body)
	}
}

func TestNewDropsEmptyErrors(t *testing.T) {
	if got := New(http.StatusBadRequest, CodeBadRequest, "Invalid request.", "", "").Response().Errors; got != nil {
		t.Fatalf("errors = %#v, want nil", got)
	}
}

func TestErrorUnwrapsCause(t *testing.T) {
	cause := errors.New("database unavailable")
	if !errors.Is(Wrap(cause, http.StatusBadRequest, CodeBadRequest, "Invalid request body."), cause) {
		t.Fatal("wrapped cause not found")
	}
}
