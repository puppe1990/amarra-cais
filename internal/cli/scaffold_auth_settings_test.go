package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// #386: logged-in password change is part of g auth / new, not a hand-rolled
// /config page. Success revokes sessions like ResetPasswordWithToken.
func TestScaffoldNewApp_includesPasswordSettings(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "settingsapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "settingsapp",
		ModulePath: "github.com/puppe1990/settingsapp",
	}, false, false); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		"internal/handlers/auth_settings.go",
		"internal/handlers/auth_settings_test.go",
		"web/templates/pages/settings.html",
	} {
		if _, err := os.Stat(filepath.Join(appDir, path)); err != nil {
			t.Errorf("missing %s: %v", path, err)
		}
	}

	storeBody := scaffoldRead(t, appDir, "internal/store/store.go")
	for _, needle := range []string{
		"FindUser(id int64)",
		"ChangePassword(id int64, passwordHash string)",
		`DELETE FROM sessions WHERE user_id = ?`,
	} {
		if !strings.Contains(storeBody, needle) {
			t.Errorf("store.go missing %q", needle)
		}
	}

	routes := scaffoldRead(t, appDir, "internal/app/routes.go")
	for _, needle := range []string{
		`middleware.RequireAuthFunc("/login", auth.Settings)`,
		`auth.SettingsPost`,
		`"/settings"`,
	} {
		if !strings.Contains(routes, needle) {
			t.Errorf("routes.go missing %q", needle)
		}
	}

	handler := scaffoldRead(t, appDir, "internal/handlers/auth_settings.go")
	for _, needle := range []string{
		"func (h *AuthHandler) Settings(",
		"func (h *AuthHandler) SettingsPost(",
		"session.VerifyPassword",
		"ChangePassword",
		"session.SignOut",
		`"/login"`,
	} {
		if !strings.Contains(handler, needle) {
			t.Errorf("auth_settings.go missing %q", needle)
		}
	}

	page := scaffoldRead(t, appDir, "web/templates/pages/settings.html")
	for _, needle := range []string{
		`action="/settings"`,
		`data-amarra-skip`,
		`name="current_password"`,
		`name="password_confirmation"`,
		`<.password`,
		`min-h-[calc(100svh-11rem)]`,
	} {
		if !strings.Contains(page, needle) {
			t.Errorf("settings.html missing %q", needle)
		}
	}

	en := scaffoldRead(t, appDir, "internal/i18n/en.go")
	pt := scaffoldRead(t, appDir, "internal/i18n/pt.go")
	for _, key := range []string{`"settings.title"`, `"settings.submit"`, `"settings.current_password"`} {
		if !strings.Contains(en, key) {
			t.Errorf("en catalog missing %s", key)
		}
		if !strings.Contains(pt, key) {
			t.Errorf("pt catalog missing %s", key)
		}
	}

	tests := scaffoldRead(t, appDir, "internal/handlers/auth_settings_test.go")
	for _, needle := range []string{
		"TestAuth_Settings_anonymousRedirectsToLogin",
		"TestAuth_SettingsPost_wrongPassword",
		"TestAuth_SettingsPost_successSignsOut",
	} {
		if !strings.Contains(tests, needle) {
			t.Errorf("auth_settings_test.go missing %s", needle)
		}
	}

	storeTest := scaffoldRead(t, appDir, "internal/store/password_reset_test.go")
	if !strings.Contains(storeTest, "TestChangePassword_revokesOnlyThatUsersSessions") {
		t.Error("password_reset_test.go missing ChangePassword session revocation test")
	}
}

func TestScaffoldAuth_addsPasswordSettingsOnBlankApp(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "blanksettings")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "blanksettings",
		ModulePath: "github.com/puppe1990/blanksettings",
	}, false, true); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldAuth(appDir, scaffoldData{
		AppName:    "blanksettings",
		ModulePath: "github.com/puppe1990/blanksettings",
	}, false); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(appDir, "internal/handlers/auth_settings.go")); err != nil {
		t.Fatalf("g auth must write auth_settings.go: %v", err)
	}
	routes := scaffoldRead(t, appDir, "internal/app/routes.go")
	if !strings.Contains(routes, `"/settings"`) {
		t.Errorf("g auth routes missing /settings:\n%s", routes)
	}
	storeBody := scaffoldRead(t, appDir, "internal/store/store.go")
	if !strings.Contains(storeBody, "ChangePassword(") {
		t.Errorf("g auth store missing ChangePassword:\n%s", storeBody)
	}
}
