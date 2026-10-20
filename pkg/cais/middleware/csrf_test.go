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

func TestCSRF_multipartBody_overLimit_returns413AndSkipsHandler(t *testing.T) {
	// #221: an oversized upload without a valid token must be rejected before
	// the handler runs and before net/http spills multipart temp files.
	called := false
	cfg := cais.Config{MaxBodyBytes: 512}
	h := CSRF(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("blob", strings.Repeat("x", 8<<10)); err != nil {
		t.Fatal(err)
	}
	if err := writer.WriteField(csrf.FormField, "secret"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: "secret"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if called {
		t.Fatal("handler must not be called for oversized body")
	}
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rr.Code)
	}
}

func TestCSRF_jsonBody_overLimit_returns413(t *testing.T) {
	called := false
	cfg := cais.Config{MaxBodyBytes: 64}
	h := CSRF(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	raw := []byte(`{"` + csrf.FormField + `":"secret","blob":"` + strings.Repeat("x", 200) + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: csrf.CookieName, Value: "secret"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if called {
		t.Fatal("handler must not be called for oversized body")
	}
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", rr.Code)
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

// #342: a crawler file has no form, and a `Set-Cookie` on it is noise the
// crawler reads as "not a public resource" — Google Search Console answers
// "Não foi possível buscar o sitemap" for a sitemap that sets a session cookie.
func TestCSRF_crawlerFiles_skipCookieInEveryEnv(t *testing.T) {
	for _, path := range []string{"/sitemap.xml", "/robots.txt", "/llms.txt"} {
		for _, cfg := range []cais.Config{{}, {Env: "production"}} {
			called := false
			h := CSRF(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			}))

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if !called {
				t.Errorf("%s (env %q): handler not called; status = %d", path, cfg.Env, rr.Code)
			}
			if got := rr.Header().Get("Set-Cookie"); got != "" {
				t.Errorf("%s (env %q): Set-Cookie = %q, want none", path, cfg.Env, got)
			}
		}
	}
}

// The crawler skip is a closed set: the app's own XML route keeps the
// double-submit cookie (a generic *.xml skip would exempt authenticated feeds).
func TestCSRF_otherXMLPath_stillSetsCookie(t *testing.T) {
	h := CSRF(cais.Config{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/feed.xml", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("Set-Cookie"); !strings.Contains(got, csrf.CookieName) {
		t.Errorf("Set-Cookie = %q, want the %s cookie", got, csrf.CookieName)
	}
}

// Regression guards: the page keeps the cookie and the unsafe method keeps the
// check, so the crawler skip cannot quietly exempt ordinary traffic.
func TestCSRF_pageAndUnsafeMethodKeepTheirBehavior(t *testing.T) {
	h := CSRF(cais.Config{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rr.Header().Get("Set-Cookie"); !strings.Contains(got, csrf.CookieName) {
		t.Errorf("GET /: Set-Cookie = %q, want the %s cookie", got, csrf.CookieName)
	}

	h = CSRF(cais.Config{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without a token")
	}))
	req := httptest.NewRequest(http.MethodPost, "/lista", strings.NewReader("nome=a"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("POST /lista without token: status = %d, want 403", rr.Code)
	}
}

// #364: the /api/ prefix is a naming convention, not proof of the auth model.
// A session-authenticated route under it must keep the double-submit check —
// a cross-site form can hit it with the victim's cookie and no token.
func TestCSRF_apiPrefixFormPostWithoutToken_rejects(t *testing.T) {
	h := CSRF(cais.Config{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/account", strings.NewReader("name=changed"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://other.example")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rr.Code)
	}
}

// #364: a JSON POST to /api/ without a Bearer is not a machine client either.
// It must present the double-submit token instead of riding the prefix.
func TestCSRF_apiPrefixJSONWithoutBearer_rejects(t *testing.T) {
	h := CSRF(cais.Config{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/projetos", strings.NewReader(`{"uc":"3004512"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rr.Code)
	}
}

// #331/#364: the Bearer skip survives, but only for the token AdminAuth would
// accept. A stray header must not turn a session route into a CSRF bypass.
func TestCSRF_adminBearerToken_skips(t *testing.T) {
	called := false
	h := CSRF(cais.Config{AdminToken: "hs_test"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/projetos", strings.NewReader(`{"uc":"1"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer hs_test")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Fatalf("handler not called for valid Bearer request; status = %d", rr.Code)
	}
}

// #364: curl -d defaults to form encoding, so a validated Bearer must still
// skip on form content. Only an unvalidated header is rejected.
func TestCSRF_adminBearerToken_skipsFormBody(t *testing.T) {
	called := false
	h := CSRF(cais.Config{AdminToken: "hs_test"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/projetos", strings.NewReader("uc=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer hs_test")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if !called {
		t.Fatalf("handler not called for valid Bearer form request; status = %d", rr.Code)
	}
}

// #364: an invalid Bearer must not exempt the request. Otherwise a same-site
// script could attach any Bearer value and fall back to the session cookie.
func TestCSRF_invalidBearerToken_doesNotSkip(t *testing.T) {
	h := CSRF(cais.Config{AdminToken: "hs_test"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/account", strings.NewReader(`{"name":"changed"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer wrong")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rr.Code)
	}
}

// #364: apps with their own API token declare the prefix and verifier
// explicitly. Only that prefix skips, and only when the verifier accepts the
// token — form content is fine once the token is validated (curl -d clients).
func TestCSRF_bearerAPI_exemptsOnlyRegisteredPrefix(t *testing.T) {
	verifyHook := func(r *http.Request) bool {
		return r.Header.Get("Authorization") == "Bearer hook-secret"
	}
	h := CSRF(cais.Config{}, BearerAPI("/hooks/", verifyHook))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	cases := []struct {
		name   string
		path   string
		auth   string
		form   bool
		status int
	}{
		{name: "registered prefix", path: "/hooks/kiwify", auth: "Bearer hook-secret", status: http.StatusNoContent},
		{name: "other prefix", path: "/other/kiwify", auth: "Bearer hook-secret", status: http.StatusForbidden},
		{name: "wrong token", path: "/hooks/kiwify", auth: "Bearer wrong", status: http.StatusForbidden},
		{name: "no token", path: "/hooks/kiwify", status: http.StatusForbidden},
		{name: "form body with valid token", path: "/hooks/kiwify", auth: "Bearer hook-secret", form: true, status: http.StatusNoContent},
		{name: "form body with wrong token", path: "/hooks/kiwify", auth: "Bearer wrong", form: true, status: http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var body io.Reader = strings.NewReader(`{"event":"paid"}`)
			contentType := "application/json"
			if tc.form {
				body = strings.NewReader("event=paid")
				contentType = "application/x-www-form-urlencoded"
			}
			req := httptest.NewRequest(http.MethodPost, tc.path, body)
			req.Header.Set("Content-Type", contentType)
			if tc.auth != "" {
				req.Header.Set("Authorization", tc.auth)
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != tc.status {
				t.Errorf("status = %d, want %d", rr.Code, tc.status)
			}
		})
	}
}

// #331/#364: a classic HTML form keeps its CSRF protection when the Bearer is
// not valid. Form posts are browser traffic, and a cross-site form cannot
// attach Authorization; that header there is noise, not a machine client.
func TestCSRF_formPostWithStrayBearer_stillRejects(t *testing.T) {
	h := CSRF(cais.Config{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("email=a%40b.c"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Bearer stray")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rr.Code)
	}
}
