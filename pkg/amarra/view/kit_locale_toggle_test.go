package view

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func TestKit_localeToggleMarksRegionQualifiedLocale(t *testing.T) {
	// #209: LOCALE=pt-BR / en-US must still mark the matching language button.
	cases := []struct {
		current, wantPressed string
	}{
		{"pt-BR", "pt"},
		{"pt_BR", "pt"},
		{"en-US", "en"},
		{"pt", "pt"},
		{"en", "en"},
	}
	for _, tc := range cases {
		t.Run(tc.current, func(t *testing.T) {
			page := fmt.Sprintf(`{{ define "content" }}<.locale-toggle current="%s" />{{ end }}`, tc.current)
			fsys := fstest.MapFS{
				"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
				"pages/home.html":  &fstest.MapFile{Data: []byte(page)},
			}
			rec, err := Load(fsys, nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{
				Layout: "app", Name: "home",
				Data: map[string]any{"CSRFToken": "tok"},
			}, cais.Config{})
			body := rr.Body.String()
			if strings.Count(body, `aria-pressed="true"`) != 1 {
				t.Fatalf("current=%q: want one aria-pressed, body=%s", tc.current, body)
			}
			enFormEnd := strings.Index(body, `value="pt"`)
			if enFormEnd < 0 {
				t.Fatalf("current=%q: missing PT form, body=%s", tc.current, body)
			}
			enPressed := strings.Contains(body[:enFormEnd], `aria-pressed="true"`)
			ptPressed := strings.Contains(body[enFormEnd:], `aria-pressed="true"`)
			if tc.wantPressed == "en" && !enPressed {
				t.Errorf("current=%q: EN should be pressed, body=%s", tc.current, body)
			}
			if tc.wantPressed == "pt" && !ptPressed {
				t.Errorf("current=%q: PT should be pressed, body=%s", tc.current, body)
			}
		})
	}
}

func TestKit_localeToggleRendersWhenCurrentMissing(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.locale-toggle />{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{
		Layout: "app", Name: "home", Data: map[string]any{"CSRFToken": "tok"},
	}, cais.Config{})
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), `value="en"`) {
		t.Errorf("missing EN form: %s", rr.Body.String())
	}
}

func TestKit_localeToggleRendersLocalesAttr(t *testing.T) {
	// #300: extra catalogs (es, zh) need a button; locales= on the tag
	// lists them. Default without the attr stays en|pt.
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html": &fstest.MapFile{Data: []byte(
			`{{ define "content" }}<.locale-toggle current="es" locales="en,pt,es" />{{ end }}`,
		)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{
		Layout: "app", Name: "home",
		Data: map[string]any{"CSRFToken": "tok"},
	}, cais.Config{})
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	for _, want := range []string{
		`value="en"`,
		`value="pt"`,
		`value="es"`,
		">ES<",
		`data-amarra-skip`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("locale-toggle missing %q in %s", want, body)
		}
	}
	if strings.Count(body, `aria-pressed="true"`) != 1 {
		t.Fatalf("want one aria-pressed, body=%s", body)
	}
	esAt := strings.Index(body, `value="es"`)
	if esAt < 0 || !strings.Contains(body[esAt:], `aria-pressed="true"`) {
		t.Errorf("ES should be pressed, body=%s", body)
	}
}

func TestKit_localeToggleRendersPageLocales(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.locale-toggle current="zh" />{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{
		Layout: "app", Name: "home",
		Data: map[string]any{"CSRFToken": "tok", "Locales": []string{"en", "zh"}},
	}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, `value="zh"`) || strings.Contains(body, `value="pt"`) {
		t.Errorf("page Locales should drive buttons, body=%s", body)
	}
}

func TestKit_localeTogglePostsToLocale(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.locale-toggle current="pt" />{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{
		Layout: "app", Name: "home",
		Data: map[string]any{"CSRFToken": "tok"},
	}, cais.Config{})
	body := rr.Body.String()
	for _, want := range []string{
		`action="/locale"`,
		`name="csrf_token"`,
		`value="tok"`,
		`name="locale"`,
		`value="en"`,
		`value="pt"`,
		`aria-pressed="true"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("locale-toggle missing %q in %s", want, body)
		}
	}
}
