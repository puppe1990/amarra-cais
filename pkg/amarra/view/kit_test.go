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
		"components/form.html":   &fstest.MapFile{Data: []byte(`<form action="{{ .Action }}" method="{{ .Method }}">{{ .Inner }}</form>`)},
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
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in %s", want, body)
		}
	}
	if strings.Contains(body, `data-amarra-drive`) {
		t.Errorf("no-op drive attr leaked into form: %s", body)
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

func TestKit_passwordRendersInputAndEyeToggle(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.password name="password" autocomplete="current-password" error="too short" />{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	body := rr.Body.String()
	for _, want := range []string{
		`type="password"`,
		`name="password"`,
		`autocomplete="current-password"`,
		`amarra-hook="password"`,
		`data-amarra-password-for="#password"`,
		`data-amarra-password-icon="show"`,
		`data-amarra-password-icon="hide"`,
		`aria-pressed="false"`,
		`aria-label="Show password"`,
		`too short`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("password kit missing %q in %s", want, body)
		}
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

func TestKit_statCard(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.stat label="Spend" value="R$ 12.00" />{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, "Spend") || !strings.Contains(body, "R$ 12.00") {
		t.Errorf("stat missing label/value: %s", body)
	}
}

func TestKit_statCardWithHrefAndDelta(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.stat label="Spend" value="R$ 12.00" href="/spend" delta="-8%" hint="vs last month" />{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	body := rr.Body.String()
	for _, want := range []string{`href="/spend"`, "-8%", "vs last month"} {
		if !strings.Contains(body, want) {
			t.Errorf("stat missing %q in %s", want, body)
		}
	}
}

func TestKit_emptyStateWithButtonSlot(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.empty title="No items"><.button type="button">Create</.button></.empty>{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, "No items") || !strings.Contains(body, "Create") {
		t.Errorf("empty missing title or slot: %s", body)
	}
}

func TestKit_emptyStateWithActionLink(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.empty title="No items" href="/items/new" action="New item">Add your first one.</.empty>{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	body := rr.Body.String()
	for _, want := range []string{`href="/items/new"`, "New item", "Add your first one."} {
		if !strings.Contains(body, want) {
			t.Errorf("empty missing %q in %s", want, body)
		}
	}
}

func tableCols() []map[string]any {
	return []map[string]any{
		{"Field": "name", "Label": "Name", "Sortable": true},
		{"Field": "created_at", "Label": "Created", "Sortable": false},
	}
}

func TestKit_tableSortableHeaders(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.table cols="{{ .Cols }}" sort="name" dir="asc">{{ range .Items }}<tr><td>{{ .Title }}</td></tr>{{ end }}</.table>{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{
		Layout: "app", Name: "home",
		Data: map[string]any{
			"Cols":  tableCols(),
			"Items": []map[string]any{{"Title": "First"}},
		},
	}, cais.Config{})
	body := rr.Body.String()
	for _, want := range []string{
		// Toggle: current sort is name asc, header links to desc.
		`href="?sort=name&dir=desc"`,
		`aria-sort="ascending"`,
		"First",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("table missing %q in %s", want, body)
		}
	}
	if strings.Contains(body, "?sort=created_at") {
		t.Errorf("non-sortable column must not emit a sort link: %s", body)
	}
}

func TestKit_tableTogglesDirOnDesc(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.table cols="{{ .Cols }}" sort="name" dir="desc"></.table>{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{"Cols": tableCols()}}, cais.Config{})
	body := rr.Body.String()
	for _, want := range []string{`href="?sort=name&dir=asc"`, `aria-sort="descending"`} {
		if !strings.Contains(body, want) {
			t.Errorf("table missing %q in %s", want, body)
		}
	}
}

func TestKit_tableFrameTargetsSortLinks(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.table cols="{{ .Cols }}" sort="name" dir="asc" frame="list"></.table>{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{"Cols": tableCols()}}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, `data-amarra-frame="list"`) {
		t.Errorf("table sort links missing frame target: %s", body)
	}
}

func TestKit_filtersGetForm(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.filters action="/items" clear="/items"><.input name="q" label="Search" value="{{ .Q }}" /></.filters>{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{"Q": "copper"}}, cais.Config{})
	body := rr.Body.String()
	for _, want := range []string{
		`action="/items"`,
		`method="get"`,
		`name="q"`,
		`value="copper"`,
		`href="/items"`, // clear link back to the index with no query
		"Clear",
		"Filter",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("filters missing %q in %s", want, body)
		}
	}
	// GET forms carry no CSRF field — only state-changing methods validate it.
	if strings.Contains(body, `name="csrf_token"`) {
		t.Errorf("filters GET form must not include a csrf_token field: %s", body)
	}
}

func TestKit_filtersFrameAndCustomLabel(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.filters action="/items" frame="list" submit="Buscar"></.filters>{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	body := rr.Body.String()
	for _, want := range []string{`data-amarra-frame="list"`, "Buscar"} {
		if !strings.Contains(body, want) {
			t.Errorf("filters missing %q in %s", want, body)
		}
	}
}

func TestKit_modalRendersNativeDialogTarget(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.modal id="confirm">Are you sure?</.modal>{{ end }}`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]any{}}, cais.Config{})
	body := rr.Body.String()
	for _, want := range []string{
		`<dialog data-amarra-dialog-target`,
		`id="confirm"`,
		"Are you sure?",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("modal missing %q in %s", want, body)
		}
	}
}

func TestShippedComponents_includesKitStems(t *testing.T) {
	got := ShippedComponents()
	for _, stem := range []string{"button", "form", "input", "flash", "nav", "pagination", "modal", "select", "textarea", "checkbox", "password", "stat", "empty", "table", "filters"} {
		if _, ok := got[stem]; !ok {
			t.Errorf("missing shipped component %s", stem)
		}
	}
}
