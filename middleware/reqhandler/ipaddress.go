package reqhandler

import (
	"net/http"
	"strings"
)

const headerXForwardedFor = "X-Forwarded-For"

// RequestIPAddress returns the IP address of the request.
// It first checks if the `X-Forwarded-For` header is set.
// If not, it falls back to using the `RemoteAddr` field.
//
// Parameters:
//   - request: The HTTP request
//
// Returns:
//   - The IP address of the request
func RequestIPAddress(request *http.Request) string {
	forwarded := request.Header.Get(headerXForwardedFor)
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IP addresses; take the first one
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Fall back to using RemoteAddr if X-Forwarded-For is not available
	ip := request.RemoteAddr
	// RemoteAddr contains IP:Port, so split by ':' and take the first part
	if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
		return ip[:colonIndex]
	}
	return ip
}
