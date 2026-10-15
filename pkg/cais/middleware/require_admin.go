package middleware

import (
	"database/sql"
	"net/http"

	"github.com/puppe1990/amarra-cais/pkg/cais/session"
)

// RequireAdmin checks the persisted role on every request so demotion takes
// effect immediately. Missing schemas and database failures deny access (#350).
func RequireAdmin(loginURL string, db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return RequireAuth(loginURL)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, _ := session.UserID(r)
			var admin bool
			err := db.QueryRowContext(r.Context(), "SELECT is_admin FROM users WHERE id = ?", id).Scan(&admin)
			if err != nil || !admin {
				http.Error(w, "Administrator permission required", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}
