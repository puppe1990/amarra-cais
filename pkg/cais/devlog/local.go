package devlog

import (
	"net/http"

	"github.com/puppe1990/amarra-cais/pkg/cais/netutil"
)

// LocalOnly restricts /logs to direct loopback connections — the buffer
// contains SQL args and request paths. Development convenience, not auth;
// never expose this endpoint on a public interface.
func LocalOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsLoopback(r) {
			http.Error(w, "logs only available on localhost", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// IsLoopback reports whether the request reached the process directly over
// loopback. The forwarding-header contract is shared with jobsui and
// devreload so a same-host proxy cannot reopen any local-only viewer (#365).
func IsLoopback(r *http.Request) bool {
	return netutil.IsLoopbackRequest(r)
}
