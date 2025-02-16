package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware(t *testing.T) {
	allowedOrigins := []string{"http://example.com"}
	allowedMethods := []string{"GET", "POST"}
	allowedHeaders := []string{"Authorization"}

	// Create a mock handler to be wrapped by the middleware
	mockHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	)

	// Create a request with an allowed origin
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")

	// Create a response recorder to capture the response
	w := httptest.NewRecorder()

	// Wrap the handler with the middleware
	handler := Middleware(
		allowedOrigins,
		allowedMethods,
		allowedHeaders,
	)(mockHandler)
	handler.ServeHTTP(w, req)

	// Check that the CORS headers were set correctly
	h := w.Header()
	assert.Equal(t, "http://example.com", h.Get(headerAllowOrigin))
	assert.Equal(t, "GET,POST", h.Get(headerAllowMethods))
	assert.Equal(t, "Content-Type,Authorization", h.Get(headerAllowHeaders))
	assert.Equal(t, "true", h.Get(headerAllowCredentials))

	// Check that the status code is OK
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCORSMiddleware_NotAllowedOrigin(t *testing.T) {
	// Define test data for allowed origins, methods, and headers
	allowedOrigins := []string{"http://example.com"}
	allowedMethods := []string{"GET", "POST"}
	allowedHeaders := []string{"Authorization"}

	// Create a mock handler to be wrapped by the middleware
	mockHandler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	)

	// Create a request with a non-allowed origin
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://notallowed.com")

	// Create a response recorder to capture the response
	w := httptest.NewRecorder()

	// Wrap the handler with the middleware
	handler := Middleware(
		allowedOrigins,
		allowedMethods,
		allowedHeaders,
	)(mockHandler)
	handler.ServeHTTP(w, req)

	// Check that the CORS headers are not set for the non-allowed origin
	h := w.Header()
	assert.Equal(t, "", h.Get(headerAllowOrigin))

	// Check that other headers are set correctly
	assert.Equal(t, "GET,POST", h.Get(headerAllowMethods))
	assert.Equal(t, "Content-Type,Authorization", h.Get(headerAllowHeaders))
	assert.Equal(t, "true", h.Get(headerAllowCredentials))

	// Check that the status code is OK
	assert.Equal(t, http.StatusOK, w.Code)
}
