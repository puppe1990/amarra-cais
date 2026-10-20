package netutil

import (
	"net/http/httptest"
	"testing"
)

// #365: RemoteAddr alone cannot prove a direct loopback connection. The
// documented deploy runs Caddy on the same host, so proxied internet traffic
// also arrives as 127.0.0.1; any forwarding header means the hop went through a
// proxy and the request is not local even when the socket is.
func TestIsLoopbackRequest(t *testing.T) {
	cases := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       bool
	}{
		{name: "ipv4 loopback", remoteAddr: "127.0.0.1:1234", want: true},
		{name: "ipv6 loopback", remoteAddr: "[::1]:1234", want: true},
		{name: "loopback without port", remoteAddr: "127.0.0.1", want: true},
		{name: "public ipv4", remoteAddr: "203.0.113.1:1234", want: false},
		{name: "public ipv6", remoteAddr: "[2001:db8::1]:1234", want: false},
		{name: "hostname", remoteAddr: "localhost:1234", want: false},
		{
			name:       "forwarded",
			remoteAddr: "127.0.0.1:1234",
			headers:    map[string]string{"Forwarded": "for=203.0.113.1;proto=https"},
			want:       false,
		},
		{
			name:       "x-real-ip",
			remoteAddr: "127.0.0.1:1234",
			headers:    map[string]string{"X-Real-IP": "203.0.113.1"},
			want:       false,
		},
		{
			name:       "x-forwarded-for",
			remoteAddr: "127.0.0.1:1234",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.1"},
			want:       false,
		},
		{
			name:       "x-forwarded-proto lowercase",
			remoteAddr: "127.0.0.1:1234",
			headers:    map[string]string{"x-forwarded-proto": "https"},
			want:       false,
		},
		{
			name:       "x-forwarded-host",
			remoteAddr: "127.0.0.1:1234",
			headers:    map[string]string{"X-Forwarded-Host": "app.example.com"},
			want:       false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tc.remoteAddr
			for name, value := range tc.headers {
				req.Header.Set(name, value)
			}
			if got := IsLoopbackRequest(req); got != tc.want {
				t.Fatalf("IsLoopbackRequest(%q, %v) = %v, want %v", tc.remoteAddr, tc.headers, got, tc.want)
			}
		})
	}
}
