package cli

const tplAuthTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"

	"{{.ModulePath}}/internal/store"
)

func newAuthHandler(t *testing.T) (*AuthHandler, store.Store) {
	t.Helper()
	s := setupTestStore(t)
	h := NewAuthHandler(setupTestViews(t), s, testSite(), s.Sessions(), cais.Config{}, i18n.DefaultCatalog())
	return h, s
}

func TestAuth_Login_redirectsWhenAuthenticated(t *testing.T) {
	h, s := newAuthHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	req = session.WithUserID(req, 1)
	rr := httptest.NewRecorder()
	h.Login(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	_ = s
}

func TestAuth_LoginPost_invalidCredentials(t *testing.T) {
	h, _ := newAuthHandler(t)

	form := url.Values{"email": {"nobody@example.com"}, "password": {"wrong"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.LoginPost(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid email or password") {
		t.Errorf("missing credentials error, got: %s", rr.Body.String())
	}
}

func TestAuth_LoginPost_validCredentials_redirects(t *testing.T) {
	s, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	h := NewAuthHandler(setupTestViews(t), s, testSite(), s.Sessions(), cais.Config{}, i18n.DefaultCatalog())

	form := url.Values{"email": {"demo@example.com"}, "password": {"password"}}
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.LoginPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303, body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Location") != "/dashboard" {
		t.Errorf("Location = %q, want /dashboard", rr.Header().Get("Location"))
	}
	foundFlash := false
	for _, c := range rr.Result().Cookies() {
		if c.Name == "cais_flash" && c.Value != "" {
			foundFlash = true
			break
		}
	}
	if !foundFlash {
		t.Error("login must set cais_flash cookie via flash.Set")
	}
}`

const tplAuthSignupTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"

	"{{.ModulePath}}/internal/store"
)

func newAuthHandlerForSignup(t *testing.T) (*AuthHandler, store.Store) {
	t.Helper()
	s, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	h := NewAuthHandler(setupTestViews(t), s, testSite(), s.Sessions(), cais.Config{}, i18n.DefaultCatalog())
	return h, s
}

func TestAuth_SignUpPost_createsUserAndRedirects(t *testing.T) {
	h, s := newAuthHandlerForSignup(t)

	form := url.Values{}
	form.Set("email", "signup@example.com")
	form.Set("password", "password123")
	form.Set("password_confirmation", "password123")
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.SignUpPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303, body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Location") != "/dashboard" {
		t.Errorf("Location = %q, want /dashboard", rr.Header().Get("Location"))
	}

	user, err := s.FindUserByEmail("signup@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == 0 {
		t.Fatal("user id = 0")
	}
}

func TestAuth_SignUpPost_duplicateEmail_returnsError(t *testing.T) {
	h, _ := newAuthHandlerForSignup(t)

	form := url.Values{}
	form.Set("email", "signup@example.com")
	form.Set("password", "password123")
	form.Set("password_confirmation", "password123")
	req := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.SignUpPost(rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Fatalf("first signup status = %d, want 303", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/signup", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr2 := httptest.NewRecorder()
	h.SignUpPost(rr2, req2)
	if rr2.Code != http.StatusUnprocessableEntity {
		t.Fatalf("duplicate signup status = %d, want 422", rr2.Code)
	}
	if !strings.Contains(rr2.Body.String(), "already registered") {
		t.Errorf("missing email taken error, got: %s", rr2.Body.String())
	}
}

func TestAuth_SignUp_RendersForm(t *testing.T) {
	h, _ := newAuthHandlerForSignup(t)

	req := httptest.NewRequest(http.MethodGet, "/signup", nil)
	rr := httptest.NewRecorder()
	h.SignUp(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), ` + "`" + `action="/signup"` + "`" + `) {
		t.Errorf("missing signup form, got: %s", rr.Body.String())
	}
}`

const tplAuthResetTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais"
	"github.com/puppe1990/amarra-cais/pkg/cais/i18n"
	"github.com/puppe1990/amarra-cais/pkg/cais/passwordreset"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"

	"{{.ModulePath}}/internal/store"
)

type captureNotifier struct {
	emails []string
	tokens []string
}

func (c *captureNotifier) NotifyReset(email, token string) error {
	c.emails = append(c.emails, email)
	c.tokens = append(c.tokens, token)
	return nil
}

func newAuthHandlerForReset(t *testing.T, s store.Store, notify passwordreset.Notifier) *AuthHandler {
	t.Helper()
	h := NewAuthHandler(setupTestViews(t), s, testSite(), s.Sessions(), cais.Config{AppURL: "http://localhost:8080"}, i18n.DefaultCatalog())
	h.resetNotify = notify
	return h
}

func TestAuth_ForgotPasswordPost_unknownEmail_showsSameMessage(t *testing.T) {
	s := setupTestStore(t)
	notify := &captureNotifier{}
	h := newAuthHandlerForReset(t, s, notify)

	form := url.Values{"email": {"missing@example.com"}}
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.ForgotPasswordPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	if len(notify.emails) != 0 {
		t.Fatal("should not notify for unknown email")
	}
}

func TestAuth_ForgotPasswordPost_knownEmail_notifiesAndRedirects(t *testing.T) {
	s, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	notify := &captureNotifier{}
	h := newAuthHandlerForReset(t, s, notify)

	form := url.Values{"email": {"demo@example.com"}}
	req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.ForgotPasswordPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	if len(notify.emails) != 1 || notify.emails[0] != "demo@example.com" {
		t.Fatalf("notify emails = %v", notify.emails)
	}
	if len(notify.tokens) != 1 || notify.tokens[0] == "" {
		t.Fatalf("notify tokens = %v", notify.tokens)
	}
}

func TestAuth_ForgotPassword_RendersForm(t *testing.T) {
	s := setupTestStore(t)
	h := newAuthHandlerForReset(t, s, &captureNotifier{})

	req := httptest.NewRequest(http.MethodGet, "/forgot-password", nil)
	rr := httptest.NewRecorder()
	h.ForgotPassword(rr, req)

	if !strings.Contains(rr.Body.String(), ` + "`" + `action="/forgot-password"` + "`" + `) {
		t.Errorf("missing forgot-password form, got: %s", rr.Body.String())
	}
}

func TestAuth_ResetPasswordPost_validToken_updatesPassword(t *testing.T) {
	s, err := store.NewSQLiteStore(":memory:", "development")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })

	user, err := s.FindUserByEmail("demo@example.com")
	if err != nil {
		t.Fatal(err)
	}
	token, err := s.CreatePasswordResetToken(user.ID)
	if err != nil {
		t.Fatal(err)
	}

	h := newAuthHandlerForReset(t, s, &captureNotifier{})
	form := url.Values{
		"token":                 {token},
		"password":              {"new-password-123"},
		"password_confirmation": {"new-password-123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.ResetPasswordPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303, body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Location") != "/login" {
		t.Errorf("Location = %q, want /login", rr.Header().Get("Location"))
	}

	updated, err := s.FindUserByEmail("demo@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if !session.VerifyPassword(updated.PasswordHash, "new-password-123") {
		t.Fatal("password was not updated")
	}
}

func TestAuth_ResetPasswordPost_invalidToken_rendersError(t *testing.T) {
	s := setupTestStore(t)
	h := newAuthHandlerForReset(t, s, &captureNotifier{})

	form := url.Values{
		"token":                 {"bad-token"},
		"password":              {"new-password-123"},
		"password_confirmation": {"new-password-123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	h.ResetPasswordPost(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "invalid or has expired") {
		t.Errorf("missing token error, got: %s", rr.Body.String())
	}
}`
