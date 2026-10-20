package devlog

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsLoopback(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       bool
	}{
		{name: "ipv4 loopback", remoteAddr: "127.0.0.1:1234", want: true},
		{name: "ipv6 loopback", remoteAddr: "[::1]:1234", want: true},
		{name: "public ipv4", remoteAddr: "203.0.113.1:1234", want: false},
		{name: "public ipv6", remoteAddr: "[2001:db8::1]:1234", want: false},
		{
			name:       "proxied loopback",
			remoteAddr: "127.0.0.1:1234",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.1"},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for name, value := range tt.headers {
				req.Header.Set(name, value)
			}
			if got := IsLoopback(req); got != tt.want {
				t.Fatalf("IsLoopback(%q) = %v, want %v", tt.remoteAddr, got, tt.want)
			}
		})
	}
}

// #365: the /logs buffer holds request paths and SQL args. A same-host proxy
// forwards external traffic from 127.0.0.1, so a forwarding header must 403.
func TestLocalOnly_rejectsProxiedLoopback(t *testing.T) {
	h := LocalOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not run for proxied traffic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rr.Code)
	}
}
