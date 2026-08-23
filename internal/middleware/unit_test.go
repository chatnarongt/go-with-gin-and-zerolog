package middleware_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/andybalholm/brotli"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/errs"
	"github.com/chatnarongt/go-with-gin-and-zerolog/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
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
			router.Use(middleware.ErrorHandler())
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

func TestCORS_AllowAllOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORS(middleware.CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
	}))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "*")
	}
}

func TestCORS_SpecificOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORS(middleware.CORSConfig{
		AllowedOrigins: []string{"http://example.com"},
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type"},
	}))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://example.com" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "http://example.com")
	}
	if got := w.Header().Get("Vary"); got != "Origin" {
		t.Errorf("Vary = %q, want %q", got, "Origin")
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORS(middleware.CORSConfig{
		AllowedOrigins: []string{"http://allowed.com"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	}))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://disallowed.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

func TestCORS_Preflight(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORS(middleware.CORSConfig{
		AllowedOrigins: []string{"http://example.com"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	}))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNoContent)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, OPTIONS" {
		t.Errorf("Access-Control-Allow-Methods = %q, want %q", got, "GET, POST, OPTIONS")
	}
	if got := w.Header().Get("Access-Control-Allow-Headers"); got != "Authorization, Content-Type" {
		t.Errorf("Access-Control-Allow-Headers = %q, want %q", got, "Authorization, Content-Type")
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CORS(middleware.CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET"},
		AllowedHeaders: []string{"Content-Type"},
	}))
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

func TestCompression_Gzip(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Compression(middleware.CompressionConfig{
		Encodings: []string{"gzip"},
		MinBytes:  10,
	}))
	payload := strings.Repeat("Hello Gzip Compression! ", 10)
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, payload)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected Content-Encoding gzip, got %q", w.Header().Get("Content-Encoding"))
	}
	if w.Header().Get("Vary") != "Accept-Encoding" {
		t.Fatalf("expected Vary Accept-Encoding, got %q", w.Header().Get("Vary"))
	}

	gr, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatalf("gzip.NewReader failed: %v", err)
	}
	defer gr.Close()
	decompressed, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("failed reading decompressed body: %v", err)
	}
	if string(decompressed) != payload {
		t.Fatalf("decompressed body = %q, want %q", string(decompressed), payload)
	}
}

func TestCompression_Zstd(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Compression(middleware.CompressionConfig{
		Encodings: []string{"zstd"},
		MinBytes:  10,
	}))
	payload := strings.Repeat("Hello Zstd Compression! ", 10)
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, payload)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "zstd")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "zstd" {
		t.Fatalf("expected Content-Encoding zstd, got %q", w.Header().Get("Content-Encoding"))
	}

	zr, err := zstd.NewReader(w.Body)
	if err != nil {
		t.Fatalf("zstd.NewReader failed: %v", err)
	}
	defer zr.Close()
	decompressed, err := io.ReadAll(zr)
	if err != nil {
		t.Fatalf("failed reading decompressed body: %v", err)
	}
	if string(decompressed) != payload {
		t.Fatalf("decompressed body = %q, want %q", string(decompressed), payload)
	}
}

func TestCompression_Brotli(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Compression(middleware.CompressionConfig{
		Encodings: []string{"br"},
		MinBytes:  10,
	}))
	payload := strings.Repeat("Hello Brotli Compression! ", 10)
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, payload)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "br")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "br" {
		t.Fatalf("expected Content-Encoding br, got %q", w.Header().Get("Content-Encoding"))
	}

	br := brotli.NewReader(w.Body)
	decompressed, err := io.ReadAll(br)
	if err != nil {
		t.Fatalf("failed reading decompressed body: %v", err)
	}
	if string(decompressed) != payload {
		t.Fatalf("decompressed body = %q, want %q", string(decompressed), payload)
	}
}

func TestCompression_Priority(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Compression(middleware.CompressionConfig{
		Encodings: []string{"zstd", "br", "gzip"},
		MinBytes:  10,
	}))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, strings.Repeat("Hello World! ", 10))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip, br, zstd")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "zstd" {
		t.Fatalf("expected server priority encoding zstd, got %q", w.Header().Get("Content-Encoding"))
	}
}

func TestCompression_QualityValueDisallowed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Compression(middleware.CompressionConfig{
		Encodings: []string{"zstd", "gzip"},
		MinBytes:  10,
	}))
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, strings.Repeat("Hello World! ", 10))
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "zstd;q=0, gzip;q=0.8")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected Content-Encoding gzip when zstd has q=0, got %q", w.Header().Get("Content-Encoding"))
	}
}

func TestCompression_BelowMinBytesSkipped(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Compression(middleware.CompressionConfig{
		Encodings: []string{"gzip"},
		MinBytes:  1000,
	}))
	payload := "Small payload"
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, payload)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "" {
		t.Fatalf("expected no Content-Encoding for small payload, got %q", w.Header().Get("Content-Encoding"))
	}
	if w.Body.String() != payload {
		t.Fatalf("body = %q, want %q", w.Body.String(), payload)
	}
	if w.Header().Get("Vary") != "Accept-Encoding" {
		t.Fatalf("expected Vary Accept-Encoding even when skipped by minBytes, got %q", w.Header().Get("Vary"))
	}
}

func TestCompression_SkipStatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Compression(middleware.CompressionConfig{
		Encodings: []string{"gzip"},
		MinBytes:  0,
	}))
	router.GET("/nocontent", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/nocontent", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "" {
		t.Fatalf("expected no Content-Encoding for 204 No Content, got %q", w.Header().Get("Content-Encoding"))
	}
}

func TestCompression_SkipPreCompressed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Compression(middleware.CompressionConfig{
		Encodings: []string{"zstd"},
		MinBytes:  0,
	}))
	router.GET("/precompressed", func(c *gin.Context) {
		c.Header("Content-Encoding", "gzip")
		c.String(http.StatusOK, strings.Repeat("Precompressed data ", 10))
	})

	req := httptest.NewRequest(http.MethodGet, "/precompressed", nil)
	req.Header.Set("Accept-Encoding", "zstd")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Fatalf("expected existing Content-Encoding gzip preserved, got %q", w.Header().Get("Content-Encoding"))
	}
}

func TestCompression_SkipSSEAndUpgrade(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Compression(middleware.CompressionConfig{
		Encodings: []string{"gzip"},
		MinBytes:  0,
	}))
	router.GET("/sse", func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		c.String(http.StatusOK, "data: test\n\n")
	})
	router.GET("/ws", func(c *gin.Context) {
		c.String(http.StatusOK, "websocket response")
	})

	// SSE
	reqSSE := httptest.NewRequest(http.MethodGet, "/sse", nil)
	reqSSE.Header.Set("Accept-Encoding", "gzip")
	wSSE := httptest.NewRecorder()
	router.ServeHTTP(wSSE, reqSSE)
	if wSSE.Header().Get("Content-Encoding") != "" {
		t.Fatalf("expected no Content-Encoding for SSE, got %q", wSSE.Header().Get("Content-Encoding"))
	}

	// Upgrade
	reqWS := httptest.NewRequest(http.MethodGet, "/ws", nil)
	reqWS.Header.Set("Accept-Encoding", "gzip")
	reqWS.Header.Set("Upgrade", "websocket")
	wWS := httptest.NewRecorder()
	router.ServeHTTP(wWS, reqWS)
	if wWS.Header().Get("Content-Encoding") != "" {
		t.Fatalf("expected no Content-Encoding for Upgrade request, got %q", wWS.Header().Get("Content-Encoding"))
	}
}
