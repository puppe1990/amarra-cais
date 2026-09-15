package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/puppe1990/amarra-cais/pkg/amarra/view"
)

func TestGenerateComponent_writesFile(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := filepath.Join(t.TempDir(), "app")
	if err := scaffoldNewApp(dir, scaffoldData{AppName: "app", ModulePath: "example.com/app"}, true, false); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	var buf bytes.Buffer
	if err := (&CLI{Out: &buf}).Run([]string{"g", "component", "card"}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "web/templates/components/card.html"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "{{ .Inner }}") {
		t.Errorf("component missing Inner slot: %s", body)
	}
}

func TestGenerateComponent_dryRunWritesNothing(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := filepath.Join(t.TempDir(), "app")
	if err := scaffoldNewApp(dir, scaffoldData{AppName: "app", ModulePath: "example.com/app"}, true, false); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	if err := (&CLI{Out: bytes.NewBuffer(nil)}).Run([]string{"g", "--dry-run", "component", "card"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web/templates/components/card.html")); !os.IsNotExist(err) {
		t.Error("dry-run should not create card.html")
	}
}

func TestDestroyComponent_removesFile(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := filepath.Join(t.TempDir(), "app")
	if err := scaffoldNewApp(dir, scaffoldData{AppName: "app", ModulePath: "example.com/app"}, true, false); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	if err := (&CLI{Out: bytes.NewBuffer(nil)}).Run([]string{"g", "component", "card"}); err != nil {
		t.Fatal(err)
	}
	if err := (&CLI{Out: bytes.NewBuffer(nil)}).Run([]string{"destroy", "component", "card"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web/templates/components/card.html")); !os.IsNotExist(err) {
		t.Error("expected card.html removed")
	}
}

func componentApp(t *testing.T) string {
	t.Helper()
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := filepath.Join(t.TempDir(), "app")
	if err := scaffoldNewApp(dir, scaffoldData{AppName: "app", ModulePath: "example.com/app"}, true, false); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	return dir
}

// A kit name must seed the shipped markup so the app restyles the real
// contract (aria-invalid, name, error slot) instead of recreating it (#63).
func TestGenerateComponent_kitNameSeedsShippedMarkup(t *testing.T) {
	dir := componentApp(t)
	if err := (&CLI{Out: bytes.NewBuffer(nil)}).Run([]string{"g", "component", "input"}); err != nil {
		t.Fatal(err)
	}

	body, err := os.ReadFile(filepath.Join(dir, "web/templates/components/input.html"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	if want := view.ShippedComponents()["input"]; got != want {
		t.Errorf("g component input must seed the shipped markup:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
	for _, needle := range []string{"aria-invalid", `name="{{ .Name }}"`} {
		if !strings.Contains(got, needle) {
			t.Errorf("seeded override lost the kit contract %q:\n%s", needle, got)
		}
	}
	if strings.Contains(got, "rounded-xl border border-slate-200") {
		t.Errorf("kit name must not fall back to the generic card:\n%s", got)
	}
}

// Kit stems keep hyphens: the override file must be locale-toggle.html or the
// <.locale-toggle> tag never resolves to it (#63).
func TestGenerateComponent_kitNameKeepsHyphenatedStem(t *testing.T) {
	dir := componentApp(t)
	if err := (&CLI{Out: bytes.NewBuffer(nil)}).Run([]string{"g", "component", "locale-toggle"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web/templates/components/locale-toggle.html")); err != nil {
		t.Fatalf("expected locale-toggle.html: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web/templates/components/locale_toggle.html")); err == nil {
		t.Error("snake_case file would not override the <.locale-toggle> tag")
	}
}

func TestGenerateComponent_unknownNameKeepsGenericTemplate(t *testing.T) {
	dir := componentApp(t)
	if err := (&CLI{Out: bytes.NewBuffer(nil)}).Run([]string{"g", "component", "hero"}); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "web/templates/components/hero.html"))
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	if !strings.Contains(got, "{{ .Inner }}") {
		t.Errorf("generic component missing the Inner slot:\n%s", got)
	}
	if strings.Contains(got, "aria-invalid") {
		t.Errorf("unknown names must not get kit markup:\n%s", got)
	}
}

// --list is discovery: it must work outside an app directory (#63).
func TestGenerateComponent_listShippedKit(t *testing.T) {
	t.Chdir(t.TempDir())
	var buf bytes.Buffer
	if err := (&CLI{Out: &buf}).Run([]string{"g", "component", "--list"}); err != nil {
		t.Fatalf("--list should not require an app: %v\n%s", err, buf.String())
	}
	out := buf.String()
	for _, name := range []string{"input", "form", "table", "locale-toggle", "password"} {
		if !strings.Contains(out, name) {
			t.Errorf("--list missing shipped component %q:\n%s", name, out)
		}
	}
}

func TestDestroyComponent_removesKitOverride(t *testing.T) {
	dir := componentApp(t)
	if err := (&CLI{Out: bytes.NewBuffer(nil)}).Run([]string{"g", "component", "input"}); err != nil {
		t.Fatal(err)
	}
	if err := (&CLI{Out: bytes.NewBuffer(nil)}).Run([]string{"destroy", "component", "input"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web/templates/components/input.html")); !os.IsNotExist(err) {
		t.Error("expected input.html removed")
	}
}
