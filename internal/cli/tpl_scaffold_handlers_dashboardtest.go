package cli

const tplDashboardTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/flash"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"

	appi18n "{{.ModulePath}}/internal/i18n"
)

func TestDashboardHandler_RendersHTML(t *testing.T) {
	h := NewDashboardHandler(setupTestViews(t), setupTestStore(t), testSite(), appi18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if !strings.Contains(rr.Body.String(), "Dashboard") {
		t.Errorf("missing dashboard heading, got: %s", rr.Body.String())
	}
}

func TestDashboardHandler_includesFlash(t *testing.T) {
	h := NewDashboardHandler(setupTestViews(t), setupTestStore(t), testSite(), appi18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req = flash.WithMessage(req, flash.Message{Kind: "notice", Message: "Welcome back!"})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Welcome back!") {
		t.Errorf("missing flash notice, got: %s", rr.Body.String())
	}
}

// #208: signed-in pages keep the rail and the sign-out form.
func TestDashboardHandler_signedInShowsAuthChrome(t *testing.T) {
	h := NewDashboardHandler(setupTestViews(t), setupTestStore(t), testSite(), appi18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req = session.WithUserID(req, 1)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, ` + "`" + `action="/logout"` + "`" + `) {
		t.Errorf("signed-in page should offer sign-out (#208), got: %s", body)
	}
	if !strings.Contains(body, ` + "`" + `href="/dashboard"` + "`" + `) {
		t.Errorf("signed-in page should link to /dashboard (#208), got: %s", body)
	}
}

func TestDashboardHandler_emitsOutlet(t *testing.T) {
	h := NewDashboardHandler(setupTestViews(t), setupTestStore(t), testSite(), appi18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	req.AddCookie(&http.Cookie{Name: "amarra_taps", Value: "4"})
	req = session.WithUserID(req, 1)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	body := rr.Body.String()
	if !strings.Contains(body, ` + "`" + `data-amarra-outlet="amarra-taps"` + "`" + `) {
		t.Errorf("dashboard GET must emit the taps outlet (#397), got: %s", body)
	}
	if !strings.Contains(body, ">4</template>") {
		t.Errorf("outlet inner should be the cookie count, got: %s", body)
	}
	if !strings.Contains(body, ` + "`" + `id="amarra-taps"` + "`" + `) {
		t.Errorf("layout must keep #amarra-taps in the shell, got: %s", body)
	}
}

func TestDashboardHandler_TapPostRedirects(t *testing.T) {
	h := NewDashboardHandler(setupTestViews(t), setupTestStore(t), testSite(), appi18n.DefaultCatalog(), cais.Config{})

	req := httptest.NewRequest(http.MethodPost, "/dashboard/tap", nil)
	req.AddCookie(&http.Cookie{Name: "amarra_taps", Value: "2"})
	rr := httptest.NewRecorder()
	h.TapPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/dashboard" {
		t.Errorf("Location = %q, want /dashboard", loc)
	}
	found := false
	for _, c := range rr.Result().Cookies() {
		if c.Name == "amarra_taps" && c.Value == "3" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected amarra_taps=3, cookies=%v", rr.Result().Cookies())
	}
}
`
