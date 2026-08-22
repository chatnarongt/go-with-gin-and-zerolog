package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/errs"
	"github.com/gin-gonic/gin"
)

func TestErrorHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		handler    gin.HandlerFunc
		expected   errs.Response
		statusCode int
	}{
		{
			name: "custom error",
			handler: func(c *gin.Context) {
				_ = c.Error(errs.BadRequest("Invalid request body.", "name: This field is required."))
			},
			expected: errs.Response{
				Status:  http.StatusBadRequest,
				Code:    errs.CodeBadRequest,
				Message: "Invalid request body.",
				Errors:  []string{"name: This field is required."},
			},
			statusCode: http.StatusBadRequest,
		},
		{
			name: "unknown error",
			handler: func(c *gin.Context) {
				_ = c.Error(errors.New("database unavailable"))
			},
			expected: errs.Response{
				Status:  http.StatusInternalServerError,
				Code:    errs.CodeInternalServerError,
				Message: "Internal server error.",
			},
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "panic",
			handler: func(*gin.Context) {
				panic("unexpected failure")
			},
			expected: errs.Response{
				Status:  http.StatusInternalServerError,
				Code:    errs.CodeInternalServerError,
				Message: "Internal server error.",
			},
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(ErrorHandler())
			router.GET("/", tt.handler)

			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.statusCode {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.statusCode)
			}

			var got errs.Response
			if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
			if got.Status != tt.expected.Status || got.Code != tt.expected.Code || got.Message != tt.expected.Message {
				t.Fatalf("response = %#v, want %#v", got, tt.expected)
			}
			if len(got.Errors) != len(tt.expected.Errors) {
				t.Fatalf("errors = %#v, want %#v", got.Errors, tt.expected.Errors)
			}
			for index := range got.Errors {
				if got.Errors[index] != tt.expected.Errors[index] {
					t.Errorf("errors[%d] = %#v, want %#v", index, got.Errors[index], tt.expected.Errors[index])
				}
			}
		})
	}
}
