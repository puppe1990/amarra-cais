package devlog

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func TestRegister_NotAvailableInProduction(t *testing.T) {
	r := cais.NewRouter()
	Register(r, "production", nil)

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}

func TestRegister_LocalDevelopmentShowsLogs(t *testing.T) {
	r := cais.NewRouter()
	buf := NewBuffer(100)
	_, _ = buf.Write([]byte("Started GET \"/\" for 127.0.0.1\n"))
	Register(r, "development", buf)

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Started GET") {
		t.Fatalf("body missing log line: %s", rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "Cais Logs") {
		t.Fatalf("body missing page title: %s", rr.Body.String())
	}
}

func TestRegister_BlocksNonLocalhost(t *testing.T) {
	r := cais.NewRouter()
	Register(r, "development", NewBuffer(100))

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	req.RemoteAddr = "203.0.113.1:1234"
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rr.Code)
	}
}

func TestRegister_PageDoesNotLoadHTMX(t *testing.T) {
	r := cais.NewRouter()
	buf := NewBuffer(100)
	_, _ = buf.Write([]byte("Started GET \"/\" for 127.0.0.1\n"))
	Register(r, "development", buf)

	req := httptest.NewRequest(http.MethodGet, "/logs", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	body := rr.Body.String()
	for _, leftover := range []string{"htmx.min.js", "hx-get", "hx-trigger", "HX-Request"} {
		if strings.Contains(body, leftover) {
			t.Errorf("logs page still references %q", leftover)
		}
	}
	if !strings.Contains(body, "/logs?partial=1") {
		t.Fatal("logs page should poll /logs?partial=1")
	}
}

func TestRegister_PartialQueryReturnsLogBody(t *testing.T) {
	r := cais.NewRouter()
	buf := NewBuffer(100)
	_, _ = buf.Write([]byte("Completed 200 OK\n"))
	Register(r, "development", buf)

	req := httptest.NewRequest(http.MethodGet, "/logs?partial=1", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if strings.Contains(rr.Body.String(), "<!DOCTYPE html>") {
		t.Fatal("partial response should not include full page")
	}
	if !strings.Contains(rr.Body.String(), "Completed 200 OK") {
		t.Fatalf("partial missing log: %s", rr.Body.String())
	}
}
