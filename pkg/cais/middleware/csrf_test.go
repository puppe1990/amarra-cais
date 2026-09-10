package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/csrf"
	"github.com/puppe1990/amarra-cais/pkg/cais/httpx"
)

func TestCSRF_safeMethod_setsCookie(t *testing.T) {
	called := false
	h := CSRF(cais.Config{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if csrf.TokenFromRequest(r) == "" {
			t.Error("expected token in request context")
		}
	}))

	req := httptest.NewRequest(http.MethodGet, "/contact", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler not called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	found := false
	for _, c := range rr.Result().Cookies() {
		if c.Name == csrf.CookieName {
			found = true
		}
	}
	if !found {
		t.Error("csrf cookie not set on GET")
	}
}

func TestCSRF_unsafeMethod_rejectsMissingToken(t *testing.T) {
	h := CSRF(cais.Config{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader("name=a"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rr.Code)
	}
}

func TestCSRF_unsafeMethod_acceptsValidToken(t *testing.T) {
	called := false
	h := CSRF(cais.Config{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	body := "name=a&csrf_token=secret"
	req := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: "secret"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler not called")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestCSRF_unsafeMethod_acceptsValidToken_multipart(t *testing.T) {
	// Classic HTML form with enctype="multipart/form-data" (e.g. file upload):
	// no X-CSRF-Token header, only the csrf_token form field (#26). ParseForm
	// does not read multipart bodies, so the field must come through FormValue.
	called := false
	h := CSRF(cais.Config{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("name", "a"); err != nil {
		t.Fatal(err)
	}
	if err := writer.WriteField(csrf.FormField, "secret"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/contact", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: "secret"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Fatalf("handler not called; status = %d", rr.Code)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestCSRF_jsonBody_handlerCanParseFormOrJSON(t *testing.T) {
	called := false
	h := CSRF(cais.Config{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if err := httpx.ParseFormOrJSON(r); err != nil {
			t.Errorf("handler ParseFormOrJSON after CSRF: %v", err)
		}
		if got := r.FormValue("email"); got != "json@example.com" {
			t.Errorf("email = %q", got)
		}
	}))

	raw, err := json.Marshal(map[string]string{"email": "json@example.com", csrf.FormField: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.Body = &csrfCloseGuard{Reader: bytes.NewReader(raw)}
	req.ContentLength = int64(len(raw))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: "secret"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if !called {
		t.Fatalf("handler not called; status = %d", rr.Code)
	}
}

type csrfCloseGuard struct {
	*bytes.Reader
	closed bool
}

func (c *csrfCloseGuard) Close() error {
	c.closed = true
	return nil
}

func (c *csrfCloseGuard) Read(p []byte) (int, error) {
	if c.closed {
		return 0, io.ErrClosedPipe
	}
	return c.Reader.Read(p)
}

func TestCSRF_skipsHealthAndStatic(t *testing.T) {
	called := false
	h := CSRF(cais.Config{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler not called for /health")
	}
}
