package view

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/middleware"
)

func themeFS() fs.FS {
	return fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}<!doctype html><html><head>
<script nonce="{{ .CSPNonce }}">
      try {
        if (localStorage.getItem("amarra-theme") === "light") document.documentElement.classList.add("light");
      } catch (e) {}
    </script>
</head><body><main id="amarra-main">{{ template "content" . }}</main></body></html>{{ end }}`)},
		"pages/home.html": &fstest.MapFile{Data: []byte(`{{ define "content" }}<h1>{{ .Title }}</h1>{{ end }}`)},
	}
}

func TestWrite_injectsCSPNonceIntoMapData(t *testing.T) {
	rec, err := Load(themeFS(), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(cais.WithScriptNonce(req.Context(), "test-nonce"))
	Write(rr, req, rec, Page{
		Layout: "app",
		Name:   "home",
		Data:   map[string]any{"Title": "Hi"},
	}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, `nonce="test-nonce"`) {
		t.Fatalf("missing injected nonce in %q", body)
	}
	if !strings.Contains(body, `localStorage.getItem("amarra-theme")`) {
		t.Fatalf("theme snippet missing in %q", body)
	}
}

func TestWrite_replacesStaleCSPNonce(t *testing.T) {
	rec, err := Load(themeFS(), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(cais.WithScriptNonce(req.Context(), "from-request"))
	Write(rr, req, rec, Page{
		Layout: "app",
		Name:   "home",
		Data:   map[string]any{"Title": "Hi", "CSPNonce": "stale"},
	}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, `nonce="from-request"`) {
		t.Fatalf("stale CSPNonce kept: %q", body)
	}
}

func TestWrite_cspNonceMatchesSecurityHeader(t *testing.T) {
	rec, err := Load(themeFS(), nil)
	if err != nil {
		t.Fatal(err)
	}
	h := middleware.SecurityHeaders(cais.Config{Env: "development"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Write(w, r, rec, Page{
			Layout: "app",
			Name:   "home",
			Data:   map[string]any{"Title": "Hi"},
		}, cais.Config{})
	}))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))

	csp := rr.Header().Get("Content-Security-Policy")
	script := cspDirective(csp, "script-src")
	if strings.Contains(script, "'unsafe-inline'") {
		t.Errorf("script-src still allows unsafe-inline: %q", script)
	}
	nonce := nonceFromScriptSrc(script)
	if nonce == "" {
		t.Fatalf("no nonce in script-src %q", script)
	}
	body := rr.Body.String()
	if !strings.Contains(body, `nonce="`+nonce+`"`) {
		t.Fatalf("HTML nonce does not match CSP %q in %q", nonce, body)
	}
	if !strings.Contains(body, `localStorage.getItem("amarra-theme")`) {
		t.Fatalf("theme snippet missing in %q", body)
	}
}

func cspDirective(csp, name string) string {
	for _, part := range strings.Split(csp, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, name+" ") || part == name {
			return part
		}
	}
	return ""
}

func nonceFromScriptSrc(script string) string {
	const prefix = "'nonce-"
	i := strings.Index(script, prefix)
	if i < 0 {
		return ""
	}
	rest := script[i+len(prefix):]
	j := strings.Index(rest, "'")
	if j < 0 {
		return ""
	}
	return rest[:j]
}
