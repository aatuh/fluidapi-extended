package reqhandler

import (
	"net"
	"net/http"
	"strings"
)

// requestIPAddress returns the IP address of the request.
func requestIPAddress(request *http.Request) string {
	forwarded := request.Header.Get(headerXForwardedFor)
	if forwarded != "" {
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}
	ip, _, err := net.SplitHostPort(request.RemoteAddr)
	if err != nil {
		return request.RemoteAddr
	}
	return ip
}
