package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldHandler_usesModernHandlerPattern(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "pageapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "pageapp",
		ModulePath: "github.com/puppe1990/pageapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldHandler(appDir, "about", false); err != nil {
		t.Fatal(err)
	}

	handler, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/about.go"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(handler)
	for _, needle := range []string{
		"view.Write",
		"meta.Site",
		"i18n.Catalog",
	} {
		if !strings.Contains(body, needle) {
			t.Errorf("about.go missing %q", needle)
		}
	}

	routes, err := os.ReadFile(filepath.Join(appDir, "internal/app/routes.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(routes), "handlers.NewAboutHandler(deps.Views, deps.Site, deps.Catalog, cfg)") {
		t.Error("routes.go should wire views, site, catalog, and cfg into handler")
	}
	viewData, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/viewdata.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(viewData), "meta.ForRequest") {
		t.Error("viewdata.go missing meta.ForRequest")
	}
}

func TestScaffoldResource_adminFormUsesFormHelpers(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "adminapp")
	if err := scaffoldNewApp(appDir, scaffoldData{
		AppName:    "adminapp",
		ModulePath: "github.com/puppe1990/adminapp",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	if err := scaffoldResource(appDir, "widget", resourceOpts{
		Fields: "name:string,url:url",
	}); err != nil {
		t.Fatal(err)
	}

	form, err := os.ReadFile(filepath.Join(appDir, "web/templates/pages/admin_widget_form.html"))
	if err != nil {
		t.Fatal(err)
	}
	body := string(form)
	if !strings.Contains(body, `csrfField`) && !strings.Contains(body, `fieldInput`) {
		t.Error("admin form should use HTML form helpers")
	}
	if !strings.Contains(body, `.Errors`) {
		t.Error("admin form should render errors")
	}

	admin, err := os.ReadFile(filepath.Join(appDir, "internal/handlers/admin_widgets.go"))
	if err != nil {
		t.Fatal(err)
	}
	adminBody := string(admin)
	if !strings.Contains(adminBody, "validate.FieldErrors") {
		t.Error("admin handler should use validate.FieldErrors")
	}
	if !strings.Contains(adminBody, "view.Write") {
		t.Error("admin handler should use view.Write on validation errors")
	}
}
