package cli

const tplLocaleHandler = `package handlers

import (
	"net/http"
	"net/url"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
)

// PostLocale persists the kit <.locale-toggle /> choice and redirects back.
func PostLocale(cfg cais.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		i18n.SetCookie(w, r.FormValue("locale"), cfg.CookieSecure())
		http.Redirect(w, r, localeNextPath(r), http.StatusSeeOther)
	}
}

func localeNextPath(r *http.Request) string {
	ref := r.Header.Get("Referer")
	if ref == "" {
		return "/"
	}
	u, err := url.Parse(ref)
	if err != nil || u.Path == "" {
		return "/"
	}
	if u.Host != "" && u.Host != r.Host {
		return "/"
	}
	if u.RawQuery != "" {
		return u.Path + "?" + u.RawQuery
	}
	return u.Path
}
`

const tplLocaleTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
)

func TestPostLocale_setsCookieAndRedirects(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/locale", strings.NewReader("locale=pt"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "example.com"
	req.Header.Set("Referer", "https://example.com/contact")
	PostLocale(cais.Config{}).ServeHTTP(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rr.Code)
	}
	if got := rr.Header().Get("Location"); got != "/contact" {
		t.Errorf("Location = %q, want /contact", got)
	}
	var cookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == i18n.CookieName {
			cookie = c
			break
		}
	}
	if cookie == nil || cookie.Value != "pt" {
		t.Fatalf("cais_locale = %+v, want pt", cookie)
	}
}

func TestPostLocale_rejectsOffHostReferer(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/locale", strings.NewReader("locale=en"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "example.com"
	req.Header.Set("Referer", "https://evil.example/phish")
	PostLocale(cais.Config{}).ServeHTTP(rr, req)
	if got := rr.Header().Get("Location"); got != "/" {
		t.Errorf("Location = %q, want /", got)
	}
}
`
