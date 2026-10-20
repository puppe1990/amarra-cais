package netutil

import (
	"net"
	"net/http"
	"strings"
)

// IsLoopbackRequest reports whether the request reached the process directly
// over the loopback interface. RemoteAddr alone is not enough: the documented
// deploy runs Caddy on the same host, so proxied internet requests also arrive
// as 127.0.0.1. Any forwarding header proves the hop went through a proxy, so
// only plain loopback connections (SSH tunnel, curl on the box) qualify (#365).
func IsLoopbackRequest(r *http.Request) bool {
	if hasForwardingHeaders(r) {
		return false
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func hasForwardingHeaders(r *http.Request) bool {
	if r.Header.Get("Forwarded") != "" || r.Header.Get("X-Real-IP") != "" {
		return true
	}
	for name := range r.Header {
		if strings.HasPrefix(strings.ToLower(name), "x-forwarded-") {
			return true
		}
	}
	return false
}
