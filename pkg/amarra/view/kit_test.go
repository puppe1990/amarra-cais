package view

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func TestKit_formIncludesCSRFAndInner(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html":       &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/login.html":       &fstest.MapFile{Data: []byte(`{{ define "content" }}<.form action="/login" method="post">{{ csrfField .CSRFToken }}<.input name="email" label="Email" value="{{ .Email }}" error="{{ fieldError .Errors "email" }}" /><.button type="submit">Go</.button></.form>{{ end }}`)},
		"components/form.html":   &fstest.MapFile{Data: []byte(`<form action="{{ .Action }}" method="{{ .Method }}" data-amarra-drive="true">{{ .Inner }}</form>`)},
		"components/input.html":  &fstest.MapFile{Data: []byte(`<label for="{{ .Name }}">{{ .Label }}</label><input id="{{ .Name }}" name="{{ .Name }}" value="{{ .Value }}">{{ if .Error }}<p class="err">{{ .Error }}</p>{{ end }}`)},
		"components/button.html": &fstest.MapFile{Data: []byte(`<button type="{{ .Type }}">{{ .Inner }}</button>`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{
		Layout: "app", Name: "login",
		Data: map[string]any{
			"Email":     "a@b.c",
			"CSRFToken": "tok",
			"Errors":    map[string]string{"email": "invalid"},
		},
	}, cais.Config{})
	body := rr.Body.String()
	for _, want := range []string{
		`action="/login"`,
		`name="email"`,
		`value="a@b.c"`,
		`invalid`,
		`type="submit"`,
		`data-amarra-drive="true"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in %s", want, body)
		}
	}
}

func TestLoad_shippedKitUsedWhenAppOmitsComponent(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.button type="submit">Go</.button>{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, "<button") || !strings.Contains(body, "Go") {
		t.Fatalf("shipped button missing: %q", body)
	}
	if !strings.Contains(body, `type="submit"`) {
		t.Fatalf("shipped button type missing: %q", body)
	}

	fsys["components/button.html"] = &fstest.MapFile{Data: []byte(`<button class="app-override">{{ .Inner }}</button>`)}
	rec, err = Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	body = rr.Body.String()
	if !strings.Contains(body, `class="app-override"`) {
		t.Fatalf("app component did not override shipped: %q", body)
	}
	if strings.Contains(body, "app-override") && strings.Contains(body, "bg-copper") {
		t.Fatalf("shipped classes leaked through override: %q", body)
	}
}

func TestKit_flashReadsPageData(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.flash />{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{
		Layout: "app", Name: "home",
		Data: map[string]any{"Flash": "Saved!"},
	}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, "Saved!") {
		t.Fatalf("flash did not read page .Flash: %q", body)
	}
}

func TestKit_shippedFormInjectsCSRF(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/login.html": &fstest.MapFile{Data: []byte(`{{ define "content" }}<.form action="/login" method="post"><.button type="submit">Go</.button></.form>{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{
		Layout: "app", Name: "login",
		Data: map[string]any{"CSRFToken": "tok"},
	}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, `name="csrf_token"`) || !strings.Contains(body, `value="tok"`) {
		t.Fatalf("shipped form missing csrf: %q", body)
	}
}

func TestKit_inputMarksInvalid(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.input name="email" label="Email" error="required" />{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, `aria-invalid="true"`) {
		t.Fatalf("invalid input missing aria-invalid: %q", body)
	}
}

func TestKit_selectTextareaCheckbox(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.select name="role" label="Role"><option>admin</option></.select><.textarea name="bio" label="Bio">hi</.textarea><.checkbox name="ok" label="OK" checked="checked" />{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	body := rr.Body.String()
	for _, want := range []string{`<select`, `name="role"`, `<textarea`, `name="bio"`, `hi`, `type="checkbox"`, `name="ok"`} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in %s", want, body)
		}
	}
}

func TestShippedComponents_includesKitStems(t *testing.T) {
	got := ShippedComponents()
	for _, stem := range []string{"button", "form", "input", "flash", "nav", "pagination", "modal", "select", "textarea", "checkbox"} {
		if _, ok := got[stem]; !ok {
			t.Errorf("missing shipped component %s", stem)
		}
	}
}
