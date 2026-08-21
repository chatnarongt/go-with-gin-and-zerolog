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

	tests := []struct {
		name     string
		err      error
		expected Response
	}{
		{
			name: "custom error",
			err:  BadRequest("Invalid request body.", Detail{Path: "name", Message: "This field is required."}),
			expected: Response{
				Status:  http.StatusBadRequest,
				Code:    CodeBadRequest,
				Message: "Invalid request body.",
				Errors:  []Detail{{Path: "name", Message: "This field is required."}},
			},
		},
		{
			name: "validation error",
			err:  validationErr,
			expected: Response{
				Status:  http.StatusBadRequest,
				Code:    CodeBadRequest,
				Message: "Invalid request body.",
				Errors:  []Detail{{Path: "email", Message: "This field is required."}},
			},
		},
		{
			name: "syntax error",
			err:  &json.SyntaxError{},
			expected: Response{
				Status:  http.StatusBadRequest,
				Code:    CodeBadRequest,
				Message: "Invalid request body.",
				Errors:  []Detail{{Path: "", Message: "Malformed JSON."}},
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

func TestResponseJSONOmitsErrorsWhenEmpty(t *testing.T) {
	body, err := json.Marshal(InternalServerError().Response())
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"status":500,"code":"INTERNAL_SERVER_ERROR","message":"Internal server error."}` {
		t.Fatalf("JSON = %s", body)
	}
}

func TestErrorUnwrapsCause(t *testing.T) {
	cause := errors.New("database unavailable")
	if !errors.Is(Wrap(cause, http.StatusBadRequest, CodeBadRequest, "Invalid request body."), cause) {
		t.Fatal("wrapped cause not found")
	}
}
