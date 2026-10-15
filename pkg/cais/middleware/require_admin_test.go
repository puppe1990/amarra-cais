package middleware

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/puppe1990/amarra-cais/pkg/cais/session"
)

func TestRequireAdmin_PersistedRole(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(1)
	for _, query := range []string{
		"CREATE TABLE users (id INTEGER PRIMARY KEY, is_admin INTEGER NOT NULL DEFAULT 0)",
		"INSERT INTO users (id) VALUES (1)",
		"INSERT INTO users VALUES (2, 1)",
	} {
		if _, err := db.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	protected := RequireAdmin("/login", db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	for _, method := range []string{"GET", "POST", "PUT", "DELETE"} {
		for _, tc := range []struct {
			id   int64
			want int
		}{{0, 303}, {1, 403}, {2, 204}, {99, 403}} {
			req := httptest.NewRequest(method, "/admin/items", nil)
			if tc.id != 0 {
				req = session.WithUserID(req, tc.id)
			}
			rr := httptest.NewRecorder()
			protected.ServeHTTP(rr, req)
			if rr.Code != tc.want {
				t.Fatalf("%s user %d: status %d, want %d", method, tc.id, rr.Code, tc.want)
			}
		}
	}
	if _, err := db.Exec("UPDATE users SET is_admin = 0 WHERE id = 2"); err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	protected.ServeHTTP(rr, session.WithUserID(httptest.NewRequest("GET", "/admin/items", nil), 2))
	if rr.Code != 403 {
		t.Fatalf("demoted admin status = %d", rr.Code)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	protected.ServeHTTP(rr, session.WithUserID(httptest.NewRequest("GET", "/admin/items", nil), 2))
	if rr.Code != 403 {
		t.Fatalf("database failure status = %d", rr.Code)
	}
}
