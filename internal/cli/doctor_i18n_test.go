package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func writeI18nFixture(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, body := range files {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func i18nCatalogPair(enKeys, ptKeys string) map[string]string {
	return map[string]string{
		"internal/i18n/i18n.go": `package i18n

var locales = map[string]map[string]string{
	"en": enMessages,
	"pt": ptMessages,
}
`,
		"internal/i18n/en.go": "package i18n\n\nvar enMessages = map[string]string{\n" + enKeys + "}\n",
		"internal/i18n/pt.go": "package i18n\n\nvar ptMessages = map[string]string{\n" + ptKeys + "}\n",
	}
}

func TestCheckI18n_skipWhenNoCatalogDir(t *testing.T) {
	checks := checkI18n(t.TempDir())
	if len(checks) != 1 {
		t.Fatalf("checks = %d, want 1 skip: %+v", len(checks), checks)
	}
	c := checks[0]
	if !c.OK {
		t.Fatalf("missing i18n dir must be ok skip, got %+v", c)
	}
	if !strings.Contains(c.Detail, "skipped") {
		t.Errorf("Detail = %q, want skipped", c.Detail)
	}
}

func TestCheckI18n_parityWarnsMissingPtKey(t *testing.T) {
	dir := t.TempDir()
	writeI18nFixture(t, dir, i18nCatalogPair(
		"\t\"hello\": \"Hello\",\n\t\"only.en\": \"EN\",\n",
		"\t\"hello\": \"Olá\",\n",
	))
	pt := mustI18nCheck(t, checkI18n(dir), "i18n pt")
	if pt.OK || !pt.Optional {
		t.Fatalf("missing pt key must warn, got %+v", pt)
	}
	if !strings.Contains(pt.Detail, "only.en") {
		t.Errorf("Detail = %q, want only.en", pt.Detail)
	}
	if !strings.Contains(pt.Detail, "chave") {
		t.Errorf("Detail = %q, want chave(s) ausente(s)", pt.Detail)
	}
	if !strings.Contains(pt.FixHint, "adicione ao catálogo pt") {
		t.Errorf("FixHint = %q, want adicione ao catálogo pt", pt.FixHint)
	}
}

func TestCheckI18n_templateLiteralMissingFromBothCatalogs(t *testing.T) {
	dir := t.TempDir()
	files := i18nCatalogPair(
		"\t\"hello\": \"Hello\",\n",
		"\t\"hello\": \"Olá\",\n",
	)
	files["web/templates/pages/dash.html"] = `{{ t "dashboard.filter" }}`
	writeI18nFixture(t, dir, files)

	checks := checkI18n(dir)
	for _, loc := range []string{"i18n en", "i18n pt"} {
		c := mustI18nCheck(t, checks, loc)
		if c.OK || !c.Optional {
			t.Fatalf("%s: used template key missing from both catalogs must warn, got %+v", loc, c)
		}
		if !strings.Contains(c.Detail, "dashboard.filter") {
			t.Errorf("%s Detail = %q, want dashboard.filter", loc, c.Detail)
		}
	}
}

func TestCheckI18n_goTCallMissingFromBothCatalogs(t *testing.T) {
	dir := t.TempDir()
	files := i18nCatalogPair(
		"\t\"hello\": \"Hello\",\n",
		"\t\"hello\": \"Olá\",\n",
	)
	files["internal/handlers/mail.go"] = `package handlers

func (h *Handler) details(r *http.Request) string {
	return h.t(r, "email.cta.join")
}
`
	writeI18nFixture(t, dir, files)

	checks := checkI18n(dir)
	for _, loc := range []string{"i18n en", "i18n pt"} {
		c := mustI18nCheck(t, checks, loc)
		if c.OK || !c.Optional {
			t.Fatalf("%s: Go .t() key missing from both catalogs must warn, got %+v", loc, c)
		}
		if !strings.Contains(c.Detail, "email.cta.join") {
			t.Errorf("%s Detail = %q, want email.cta.join", loc, c.Detail)
		}
	}
}

func TestCheckI18n_skipsPrintfDynamicKeys(t *testing.T) {
	dir := t.TempDir()
	files := i18nCatalogPair(
		"\t\"hello\": \"Hello\",\n",
		"\t\"hello\": \"Olá\",\n",
	)
	files["internal/handlers/mail.go"] = `package handlers

func (h *Handler) details(r *http.Request) string {
	return h.t(r, "email.%s")
}
`
	files["web/templates/pages/dash.html"] = `{{ t (printf "prefix.%s" .X) }}`
	writeI18nFixture(t, dir, files)

	assertI18nAllOK(t, checkI18n(dir))
}

func TestCheckI18n_okWhenCatalogsMatchAndKeysUsed(t *testing.T) {
	dir := t.TempDir()
	files := i18nCatalogPair(
		"\t\"hello\": \"Hello\",\n\t\"email.cta.join\": \"Join\",\n",
		"\t\"hello\": \"Olá\",\n\t\"email.cta.join\": \"Entrar\",\n",
	)
	files["internal/handlers/mail.go"] = `package handlers

func (h *Handler) details(r *http.Request) string {
	return h.t(r, "email.cta.join")
}
`
	files["web/templates/pages/hello.html"] = `{{ t "hello" }}`
	writeI18nFixture(t, dir, files)

	assertI18nAllOK(t, checkI18n(dir))
}

func TestCheckI18n_freshScaffoldOK(t *testing.T) {
	unsetCIEnv(t)
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "ok",
		ModulePath: "github.com/puppe1990/ok",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	assertI18nAllOK(t, checkI18n(dir))
}

func TestDoctor_i18nParityAppearsAsWarn(t *testing.T) {
	unsetCIEnv(t)
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := t.TempDir()
	if err := scaffoldNewApp(dir, scaffoldData{
		AppName:    "ok",
		ModulePath: "github.com/puppe1990/ok",
	}, true, false); err != nil {
		t.Fatal(err)
	}
	writeBuiltStylesCSS(t, dir)
	ptPath := filepath.Join(dir, "internal/i18n/pt.go")
	body, err := os.ReadFile(ptPath)
	if err != nil {
		t.Fatal(err)
	}
	re := regexp.MustCompile(`\t"auth\.welcome":\s+"Bem-vindo de volta",\n`)
	patched := re.ReplaceAllString(string(body), "")
	if patched == string(body) {
		t.Fatalf("could not strip auth.welcome from pt catalog:\n%s", body)
	}
	if err := os.WriteFile(ptPath, []byte(patched), 0o644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := runDoctor(&buf, dir, doctorOptions{}); err != nil {
		t.Fatalf("i18n warn must not fail doctor: %v\n%s", err, buf.String())
	}
	out := buf.String()
	if !strings.Contains(out, "[warn] i18n pt") {
		t.Errorf("expected [warn] i18n pt, got:\n%s", out)
	}
	if !strings.Contains(out, "auth.welcome") {
		t.Errorf("expected auth.welcome in warn, got:\n%s", out)
	}
	if !strings.Contains(out, "fix: adicione ao catálogo pt") {
		t.Errorf("expected pt fix hint, got:\n%s", out)
	}
}

func mustI18nCheck(t *testing.T, checks []doctorCheck, name string) doctorCheck {
	t.Helper()
	for _, c := range checks {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("missing check %q in %+v", name, checks)
	return doctorCheck{}
}

func assertI18nAllOK(t *testing.T, checks []doctorCheck) {
	t.Helper()
	if len(checks) == 0 {
		t.Fatal("expected at least one i18n check")
	}
	for _, c := range checks {
		if !c.OK {
			t.Errorf("i18n must be ok, got %+v", c)
		}
	}
}
