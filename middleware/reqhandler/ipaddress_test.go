package reqhandler

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRequestIPAddress tests the RequestIPAddress function with X-Forwarded-For
// header.
func TestRequestIPAddress_WithXForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(headerXForwardedFor, "192.0.2.1, 203.0.113.195")

	ip := RequestIPAddress(req)

	assert.Equal(t, "192.0.2.1", ip, "Expected to extract the first IP address from X-Forwarded-For header")
}

// TestRequestIPAddress tests the RequestIPAddress function with a single
// X-Forwarded-For header address.
func TestRequestIPAddress_WithSingleXForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set(headerXForwardedFor, "192.0.2.1")

	ip := RequestIPAddress(req)

	assert.Equal(t, "192.0.2.1", ip, "Expected to extract the single IP address from X-Forwarded-For header")
}

// TestRequestIPAddress tests the RequestIPAddress function with no
// X-Forwarded-For header.
func TestRequestIPAddress_WithoutXForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "192.0.2.1:12345"

	ip := RequestIPAddress(req)

	assert.Equal(t, "192.0.2.1", ip, "Expected to extract IP address from RemoteAddr")
}

// TestRequestIPAddress tests the RequestIPAddress function with no
// X-Forwarded-For header and an invalid RemoteAddr
func TestRequestIPAddress_WithoutXForwardedFor_InvalidRemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "192.0.2.1"

	ip := RequestIPAddress(req)

	assert.Equal(t, "192.0.2.1", ip, "Expected to extract IP address without port from RemoteAddr")
}

// TestRequestIPAddress tests the RequestIPAddress function with no
// X-Forwarded-For header and an empty RemoteAddr
func TestRequestIPAddress_WithEmptyXForwardedForAndRemoteAddr(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = ""

	ip := RequestIPAddress(req)

	assert.Equal(t, "", ip, "Expected empty IP address when both X-Forwarded-For and RemoteAddr are empty")
}
