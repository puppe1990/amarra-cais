package cli

const tplAuthSettingsTest = `package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/cais/session"

	"{{.ModulePath}}/internal/models"
	"{{.ModulePath}}/internal/store"
)

func createSettingsUser(t *testing.T, s store.Store) models.User {
	t.Helper()
	hash, err := session.HashPassword("password")
	if err != nil {
		t.Fatal(err)
	}
	id, err := s.CreateUser("member@example.com", hash)
	if err != nil {
		t.Fatal(err)
	}
	u, err := s.FindUser(id)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func TestAuth_Settings_anonymousRedirectsToLogin(t *testing.T) {
	h, _ := newAuthHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/settings", nil)
	rr := httptest.NewRecorder()
	h.Settings(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want 303", rr.Code)
	}
	if rr.Header().Get("Location") != "/login" {
		t.Errorf("Location = %q, want /login", rr.Header().Get("Location"))
	}
}

func TestAuth_SettingsPost_wrongPassword(t *testing.T) {
	h, s := newAuthHandler(t)
	user := createSettingsUser(t, s)
	form := url.Values{
		"current_password":      {"nope"},
		"password":              {"new-password-123"},
		"password_confirmation": {"new-password-123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = session.WithUserID(req, user.ID)
	rr := httptest.NewRecorder()
	h.SettingsPost(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422, body: %s", rr.Code, rr.Body.String())
	}
}

func TestAuth_SettingsPost_successSignsOut(t *testing.T) {
	h, s := newAuthHandler(t)
	user := createSettingsUser(t, s)
	otherHash, err := session.HashPassword("password")
	if err != nil {
		t.Fatal(err)
	}
	otherID, err := s.CreateUser("other@example.com", otherHash)
	if err != nil {
		t.Fatal(err)
	}

	token, err := s.Sessions().Create(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	otherToken, err := s.Sessions().Create(otherID)
	if err != nil {
		t.Fatal(err)
	}

	form := url.Values{
		"current_password":      {"password"},
		"password":              {"new-password-123"},
		"password_confirmation": {"new-password-123"},
	}
	req := httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req = session.WithUserID(req, user.ID)
	req.AddCookie(&http.Cookie{Name: session.DefaultCookieName, Value: token})
	rr := httptest.NewRecorder()
	h.SettingsPost(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303, body: %s", rr.Code, rr.Body.String())
	}
	if rr.Header().Get("Location") != "/login" {
		t.Errorf("Location = %q, want /login", rr.Header().Get("Location"))
	}
	if _, ok := s.Sessions().Get(token); ok {
		t.Fatal("user session must be revoked")
	}
	if _, ok := s.Sessions().Get(otherToken); !ok {
		t.Fatal("other user session must remain")
	}
	updated, err := s.FindUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !session.VerifyPassword(updated.PasswordHash, "new-password-123") {
		t.Fatal("password was not updated")
	}
	cleared := false
	for _, c := range rr.Result().Cookies() {
		if c.Name == session.DefaultCookieName && (c.MaxAge < 0 || c.Value == "") {
			cleared = true
		}
	}
	if !cleared {
		t.Fatal("success must clear session cookie")
	}
}
`
