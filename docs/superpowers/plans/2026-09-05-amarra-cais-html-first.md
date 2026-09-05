# amarra-cais HTML-first (Slice A) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create the `amarra-cais` fork of Cais v0.11 so `amarra-cais new` scaffolds an HTML-first app (Amarra Views + Drive/Frame/Stream/Hook) with zero Inertia/Svelte/Vite.

**Architecture:** New repo `github.com/puppe1990/amarra-cais` copied from Cais. Core (`pkg/cais`) stays. New `pkg/amarra` owns views (html/template + `<.component>` preprocessor), HTTP Drive/Frame, SSE Stream, and `amarra.js`. CLI binary is `amarra-cais`. Live WebSocket is a 501 stub only.

**Tech Stack:** Go 1.26, `html/template`, Idiomorph (vendored), Tailwind 3, SQLite (`modernc.org/sqlite`), esbuild, `node --test`. No gonertia, no Svelte, no Vite, no HTMX as a public API.

**Spec:** `docs/superpowers/specs/2026-09-05-amarra-cais-design.md`

**Workdir:** `/Users/matheuspuppe/Desktop/Projetos/amarra-cais` (created in Task 1). Do not implement Slice A on `puppe1990/cais` `main`.

**Out of this plan (Slice B, later):** Live hub, WS Mount/Handle, `amarra-live` reconnect, in-process broadcast. `jobsui` keeps its current internal HTML; do not rewrite it.

---

## File Map

| File                                | Responsibility                                                           |
| ----------------------------------- | ------------------------------------------------------------------------ |
| `cmd/amarra-cais/main.go`           | CLI entry (renamed from `cmd/cais`)                                      |
| `pkg/cais/boot/version.go`          | Module path `github.com/puppe1990/amarra-cais`                           |
| `pkg/amarra/view/expand.go`         | `<.component>` preprocessor                                              |
| `pkg/amarra/view/renderer.go`       | Load templates, execute layouts/pages/components                         |
| `pkg/amarra/view/write.go`          | `Write` full / Drive / Frame                                             |
| `pkg/amarra/view/helpers.go`        | `linkTo`, `render`, `renderEach`, `fieldError`                           |
| `pkg/amarra/view/kit.go`            | Parse shipped `form`/`input`/`button`/`flash`/`nav`/`pagination`/`modal` |
| `pkg/amarra/view/components/*.html` | Shipped kit sources                                                      |
| `pkg/amarra/drive.go`               | `IsDrive`, header `Amarra-Drive`                                         |
| `pkg/amarra/frame.go`               | `FrameID`, header `Amarra-Frame`                                         |
| `pkg/amarra/stream/stream.go`       | Moved from `pkg/cais/stream` + named ops                                 |
| `pkg/amarra/live/live.go`           | `Handler` → 501                                                          |
| `pkg/amarra/js/drive.mjs`           | Click/submit fetch + morph                                               |
| `pkg/amarra/js/frame.mjs`           | `<amarra-frame>` fetches                                                 |
| `pkg/amarra/js/stream.mjs`          | SSE / HTTP ops                                                           |
| `pkg/amarra/js/hook.mjs`            | CSRF header, toast, focus, optimistic                                    |
| `pkg/amarra/js/morph.mjs`           | Idiomorph wrapper                                                        |
| `pkg/amarra/js/entry.mjs`           | Boot                                                                     |
| `pkg/cais/pwa/assets/amarra.js`     | esbuild IIFE output                                                      |
| `internal/cli/*.go`                 | `amarra-cais` help, scaffold, generators, doctor, dev                    |
| `scripts/js-build.mjs`              | Bundle `entry.mjs` → `amarra.js`                                         |
| `scripts/smoke-scaffold.sh`         | `amarra-cais new` + `GET /` HTML                                         |

---

## Locked contracts

```go
const HeaderDrive = "Amarra-Drive" // value "true"
const HeaderFrame = "Amarra-Frame" // frame id

type Page struct {
	Layout string // default "app"
	Name   string // "home", "items/index"
	Frame  string // optional define name "frame:cart"
	Data   any
	Status int
}

func Write(w http.ResponseWriter, r *http.Request, rec *view.Renderer, p Page, cfg cais.Config)
```

Drive response: HTML with `#amarra-main` (full layout is fine; JS selects the node). Frame response: the frame root only. Full page (no Drive header): layout + page.

Component files use `{{ .Inner }}` for the slot and `{{ .Type }}` (Pascal attr) for attrs. Invocation: `<.button type="submit">Save</.button>` or `<.flash />`.

---

### Task 1: Fork repo, module path, CLI name

**Files:**

- Create: `/Users/matheuspuppe/Desktop/Projetos/amarra-cais` (copy of Cais)
- Modify: every `github.com/puppe1990/cais` import → `github.com/puppe1990/amarra-cais`
- Create: `cmd/amarra-cais/main.go`
- Delete: `cmd/cais/`
- Modify: `pkg/cais/boot/version.go`, `internal/cli/cli.go`, `Makefile`, `go.mod`
- Test: `internal/cli/version_test.go`, `internal/cli/cli_help_test.go`

- [ ] **Step 1: Copy the tree**

```bash
cp -a /Users/matheuspuppe/Desktop/Projetos/cais /Users/matheuspuppe/Desktop/Projetos/amarra-cais
cd /Users/matheuspuppe/Desktop/Projetos/amarra-cais
rm -rf .git
git init
git add .
git commit -m "chore: import Cais v0.11 as amarra-cais starting point"
cp docs/superpowers/specs/2026-09-05-amarra-cais-design.md docs/superpowers/specs/
cp docs/superpowers/plans/2026-09-05-amarra-cais-html-first.md docs/superpowers/plans/
```

All later steps run in `/Users/matheuspuppe/Desktop/Projetos/amarra-cais`.

- [ ] **Step 2: Write the failing CLI help test**

Replace `internal/cli/cli_help_test.go` assertions (keep the file; change needles):

```go
func TestCLI_Help_UsesAmarraCais(t *testing.T) {
	var buf bytes.Buffer
	c := &CLI{Out: &buf}
	if err := c.Run([]string{"help"}); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"amarra-cais new",
		"amarra-cais version",
		"amarra-cais g",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q", want)
		}
	}
	if strings.Contains(out, "cais new") {
		t.Error("help still documents cais new")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./internal/cli -run TestCLI_Help_UsesAmarraCais -count=1
```

Expected: FAIL — help still says `cais new`.

- [ ] **Step 4: Rewrite module + CLI name**

```bash
# go.mod module line
# then:
grep -rl 'github.com/puppe1990/cais' --include='*.go' --include='go.mod' --include='*.md' --include='*.yml' |
  xargs sed -i '' 's|github.com/puppe1990/cais|github.com/puppe1990/amarra-cais|g'
```

`pkg/cais/boot/version.go`:

```go
const modulePath = "github.com/puppe1990/amarra-cais"
```

Keep the function name `CaisVersion()` (internal); no mass-rename of `pkg/cais` types.

```bash
mkdir -p cmd/amarra-cais
mv cmd/cais/main.go cmd/amarra-cais/main.go
rmdir cmd/cais
```

`cmd/amarra-cais/main.go`:

```go
package main

import (
	"os"

	"github.com/puppe1990/amarra-cais/internal/cli"
)

func main() {
	os.Exit(cli.Main())
}
```

`internal/cli/cli.go` `Main` error prefix: `amarra-cais: %v`. `printHelp` usage lines: `amarra-cais …`. `isCaisApp` / `isCaisFramework` match `github.com/puppe1990/amarra-cais`.

`Makefile`: `BIN := bin/amarra-cais`, `go build … ./cmd/amarra-cais`, `go install ./cmd/amarra-cais`.

`internal/cli/scaffold.go`: `defaultScaffoldCaisVersion = "0.1.0"`.

- [ ] **Step 5: Run tests and commit**

```bash
go test ./internal/cli -run 'TestCLI_Help_UsesAmarraCais|TestCLI_Version' -count=1
go test ./pkg/cais/boot -count=1
```

Expected: PASS.

```bash
git add -A
git commit -m "chore: rename module and CLI to amarra-cais"
```

---

### Task 2: `<.component>` preprocessor

**Files:**

- Create: `pkg/amarra/view/expand.go`
- Test: `pkg/amarra/view/expand_test.go`

- [ ] **Step 1: Write the failing tests**

```go
package view

import (
	"strings"
	"testing"
)

func TestExpandAll_staticButton(t *testing.T) {
	components := map[string]string{
		"button": `<button type="{{ .Type }}">{{ .Inner }}</button>`,
	}
	got, err := ExpandAll(`<.button type="submit">Save</.button>`, components)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`{{ $type := "submit" }}`,
		`<button type="{{ $type }}">`,
		`Save`,
		`</button>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in %q", want, got)
		}
	}
	if strings.Contains(got, "<.button") {
		t.Errorf("left component tag in %q", got)
	}
}

func TestExpandAll_selfClosing(t *testing.T) {
	components := map[string]string{
		"flash": `<div class="flash">{{ .Inner }}</div>`,
	}
	got, err := ExpandAll(`<.flash />`, components)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "<.flash") {
		t.Fatalf("not expanded: %q", got)
	}
	if !strings.Contains(got, `<div class="flash">`) {
		t.Errorf("got %q", got)
	}
}

func TestExpandAll_nested(t *testing.T) {
	components := map[string]string{
		"form":   `<form>{{ .Inner }}</form>`,
		"button": `<button>{{ .Inner }}</button>`,
	}
	got, err := ExpandAll(`<.form><.button>Go</.button></.form>`, components)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "<.form") || strings.Contains(got, "<.button") {
		t.Fatalf("not fully expanded: %q", got)
	}
	if !strings.Contains(got, `<form>`) || !strings.Contains(got, `<button>`) || !strings.Contains(got, `Go`) {
		t.Errorf("got %q", got)
	}
}

func TestExpandAll_unknownComponent(t *testing.T) {
	_, err := ExpandAll(`<.nope />`, map[string]string{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExpandAll_dynamicAttr(t *testing.T) {
	components := map[string]string{
		"input": `<input value="{{ .Value }}">`,
	}
	got, err := ExpandAll(`<.input value="{{ .Item.Title }}" />`, components)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `{{ $value := .Item.Title }}`) {
		t.Errorf("got %q", got)
	}
}
```

Use `strings.Contains` instead of the hand-rolled helpers when implementing the test file for real.

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/amarra/view -run TestExpandAll -count=1
```

Expected: FAIL — `view` package does not exist.

- [ ] **Step 3: Write minimal implementation**

`pkg/amarra/view/expand.go`:

- Scan for `<.name …>` / `<.name … />` / `</.name>`.
- Expand innermost (or first complete) invocation until none remain.
- Attr `foo="bar"` → `{{ $foo := "bar" }}` then rewrite `{{ .Foo }}` in the component body to `{{ $foo }}` (Pascal → `$` + original attr ident).
- Attr `foo="{{ .X }}"` → `{{ $foo := .X }}`.
- Replace `{{ .Inner }}` with the slot source (empty if self-closing).
- Unknown name: `fmt.Errorf("unknown amarra component %q", name)`.

Keep files under ~200 lines; put the scanner in `expand_scan.go` if `expand.go` grows.

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./pkg/amarra/view -run TestExpandAll -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/amarra/view/expand.go pkg/amarra/view/expand_scan.go pkg/amarra/view/expand_test.go
git commit -m "feat(view): expand <.component> tags into html/template"
```

---

### Task 3: `view.Load` + `view.Write` (full / Drive / Frame)

**Files:**

- Create: `pkg/amarra/view/renderer.go`, `pkg/amarra/view/write.go`
- Create: `pkg/amarra/drive.go`, `pkg/amarra/frame.go`
- Test: `pkg/amarra/view/write_test.go`
- Modify: none of `pkg/cais/render.go` yet (keep it for jobsui/devlog)

- [ ] **Step 1: Write the failing tests**

```go
package view

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/puppe1990/amarra-cais/pkg/amarra"
	"github.com/puppe1990/amarra-cais/pkg/cais"
)

func testFS() fs.FS {
	return fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}<!doctype html><html><body><nav id="amarra-nav">n</nav><main id="amarra-main">{{ template "content" . }}</main></body></html>{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<h1>{{ .Title }}</h1>{{ define "frame:box" }}<section id="box">{{ .Title }}</section>{{ end }}{{ end }}`)},
	}
}

func TestWrite_fullPage(t *testing.T) {
	rec, err := Load(testFS(), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	Write(rr, req, rec, Page{Layout: "app", Name: "home", Data: map[string]string{"Title": "Hi"}}, cais.Config{})
	body := rr.Body.String()
	if rr.Code != 200 {
		t.Fatalf("code %d", rr.Code)
	}
	if !strings.Contains(body, `id="amarra-main"`) || !strings.Contains(body, "<h1>Hi</h1>") {
		t.Fatalf("body %q", body)
	}
}

func TestWrite_driveSelectsMain(t *testing.T) {
	rec, err := Load(testFS(), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(amarra.HeaderDrive, "true")
	Write(rr, req, rec, Page{Layout: "app", Name: "home", Data: map[string]string{"Title": "Hi"}}, cais.Config{})
	body := rr.Body.String()
	if !strings.Contains(body, `id="amarra-main"`) || !strings.Contains(body, "<h1>Hi</h1>") {
		t.Fatalf("body %q", body)
	}
}

func TestWrite_frame(t *testing.T) {
	rec, err := Load(testFS(), nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(amarra.HeaderFrame, "box")
	Write(rr, req, rec, Page{Layout: "app", Name: "home", Frame: "box", Data: map[string]string{"Title": "Hi"}}, cais.Config{})
	body := rr.Body.String()
	if strings.Contains(body, `id="amarra-nav"`) {
		t.Fatalf("frame leaked layout: %q", body)
	}
	if !strings.Contains(body, `<section id="box">Hi</section>`) {
		t.Fatalf("body %q", body)
	}
}

func TestLoad_expandsComponents(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html":         &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":          &fstest.MapFile{Data: []byte(`{{ define "content" }}<.button type="submit">Go</.button>{{ end }}`)},
		"components/button.html":   &fstest.MapFile{Data: []byte(`<button type="{{ .Type }}">{{ .Inner }}</button>`)},
	}
	rec, err := Load(fsys, nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	Write(rr, httptest.NewRequest(http.MethodGet, "/", nil), rec, Page{Layout: "app", Name: "home", Data: map[string]string{}}, cais.Config{})
	if !strings.Contains(rr.Body.String(), `<button type="submit">Go</button>`) {
		t.Fatalf("body %q", rr.Body.String())
	}
}

func TestLoad_unknownComponentFailsBoot(t *testing.T) {
	fsys := fstest.MapFS{
		"layouts/app.html": &fstest.MapFile{Data: []byte(`{{ define "app" }}{{ template "content" . }}{{ end }}`)},
		"pages/home.html":  &fstest.MapFile{Data: []byte(`{{ define "content" }}<.missing />{{ end }}`)},
	}
	if _, err := Load(fsys, nil); err == nil {
		t.Fatal("expected boot error")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/amarra/view -run 'TestWrite_|TestLoad_' -count=1
```

Expected: FAIL — `Load` / `Write` undefined.

- [ ] **Step 3: Write minimal implementation**

`pkg/amarra/drive.go`:

```go
package amarra

const HeaderDrive = "Amarra-Drive"

func IsDrive(r *http.Request) bool {
	return r.Header.Get(HeaderDrive) == "true"
}
```

`pkg/amarra/frame.go`:

```go
package amarra

const HeaderFrame = "Amarra-Frame"

func FrameID(r *http.Request) string {
	return r.Header.Get(HeaderFrame)
}
```

`Load`:

1. Read `layouts/*.html`, `pages/**/*.html` (or `pages/*.html` plus one subdirectory level for `items/index.html`), `components/*.html`.
2. `ExpandAll` every page and layout using the component map (filename stem → source). Kit components from Task 4 can be empty now; if a page has no `<.` tags, expand is a no-op.
3. Parse with `html/template` + i18n/meta/forms func maps (copy the merge from `pkg/cais/render.go` `templateFuncs`).
4. Store per-page templates.

`Write`: set `Content-Type: text/html; charset=utf-8`; if `Status != 0` use it. If `FrameID(r) != ""` or `p.Frame != ""`, execute the `frame:<id>` define only. Else execute layout `p.Layout` (default `"app"`). On execute error, use `cfg.SanitizeErrors()` the same way `httpx.writeRenderError` does.

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./pkg/amarra/... -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/amarra
git commit -m "feat(view): Load templates and Write full, Drive, and Frame HTML"
```

---

### Task 4: Kit + helpers

**Files:**

- Create: `pkg/amarra/view/components/{button,form,input,flash,nav,pagination,modal}.html`
- Create: `pkg/amarra/view/helpers.go`, `pkg/amarra/view/kit.go`
- Test: `pkg/amarra/view/kit_test.go`, `pkg/amarra/view/helpers_test.go`

- [ ] **Step 1: Write the failing tests**

```go
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

func TestLinkTo_driveAnchor(t *testing.T) {
	got := string(LinkTo("/x", "Hi"))
	if !strings.Contains(got, `href="/x"`) || !strings.Contains(got, "Hi") {
		t.Fatalf("%q", got)
	}
}
```

CSRF stays a page-level `{{ csrfField .CSRFToken }}` inside `<.form>` inner HTML. `<.form>` only emits the `<form>` tag.

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/amarra/view -run 'TestKit_|TestLinkTo_' -count=1
```

Expected: FAIL.

- [ ] **Step 3: Write kit HTML + helpers**

Shipped components (Tailwind classes aligned with current Cais forms: `rounded-lg`, `indigo-600`):

- `button.html` — `type`, `class`, `click` → optional `amarra-click`
- `form.html` — `action`, `method` default `post`, `data-amarra-drive="true"`
- `input.html` — `name`, `label`, `value`, `type` default `text`, `error`
- `flash.html` — renders `.Flash` message when present; page may pass inner empty
- `nav.html`, `pagination.html`, `modal.html` — minimal markup so `cais g resource` has targets; keep each file < 40 lines

`helpers.go` FuncMap extras (merged in `Load`):

```go
"linkTo":      LinkTo,      // template.HTML `<a href="..." data-amarra-drive="true">`
"render":      nil,         // implemented via {{ template }} — skip if not needed
"renderEach":  RenderEach,  // only if you can execute named templates; otherwise use {{ range }} in pages
"fieldError":  FieldError,  // copy pkg/cais/forms.FieldError
```

Do **not** implement a magic `render` engine if `{{ template "items/_form" . }}` already works after ParseFS. Prefer `{{ template }}`. Drop `render`/`renderEach` from FuncMap if that keeps YAGNI; keep `linkTo` + `fieldError`.

`kit.go`: `ShippedComponents() map[string]string` via `//go:embed components/*.html`. `Load` merges shipped then **app files override same name**.

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./pkg/amarra/view -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/amarra/view
git commit -m "feat(view): ship form/input/button/flash kit and linkTo helper"
```

---

### Task 5: Move stream + named ops

**Files:**

- Create: `pkg/amarra/stream/stream.go` (move `pkg/cais/stream`)
- Create: `pkg/amarra/stream/op.go`
- Test: move `pkg/cais/stream/*_test.go` and add `op_test.go`
- Modify: every importer of `pkg/cais/stream` (chat, cli templates, testdata)
- Delete: `pkg/cais/stream/` after importers compile

- [ ] **Step 1: Write the failing op test**

```go
func TestWriteOp_append(t *testing.T) {
	rr := httptest.NewRecorder()
	if err := WriteOp(rr, Op{Kind: "append", Target: "chat-history", HTML: "<p>hi</p>"}); err != nil {
		t.Fatal(err)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "event: append") || !strings.Contains(body, "data: <p>hi</p>") {
		t.Fatalf("%q", body)
	}
}
```

Keep existing `WriteEvent` tests; they must still pass on the new import path.

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/amarra/stream -run TestWriteOp_append -count=1
```

Expected: FAIL — package missing.

- [ ] **Step 3: Move package + add `Op`**

```go
type Op struct {
	Kind   string // append, prepend, replace, morph, remove, toast
	Target string
	HTML   string
}

func WriteOp(w http.ResponseWriter, op Op) error {
	return WriteEvent(w, op.Kind, encodeOp(op))
}
```

`encodeOp` for `toast` can be the message string; for DOM ops, HTML payload. `remove` HTML may be empty (`data: ` line still emitted).

Copy `RelaySSE`, `WriteEvent`, `Flush` unchanged (they already split `\n` in data lines).

Update imports with:

```bash
grep -rl 'pkg/cais/stream' --include='*.go' | xargs sed -i '' 's|pkg/cais/stream|pkg/amarra/stream|g'
```

- [ ] **Step 4: Run tests**

```bash
go test ./pkg/amarra/stream ./pkg/cais/... -count=1
```

Expected: PASS. No remaining `pkg/cais/stream`.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "feat(stream): move SSE under pkg/amarra/stream and add named ops"
```

---

### Task 6: Live stub (501)

**Files:**

- Create: `pkg/amarra/live/live.go`
- Test: `pkg/amarra/live/live_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestHandler_returns501(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/amarra/live", nil)
	Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusNotImplemented {
		t.Fatalf("code %d", rr.Code)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/amarra/live -count=1
```

Expected: FAIL.

- [ ] **Step 3: Implement**

```go
package live

func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "amarra live is not enabled in this release", http.StatusNotImplemented)
	})
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./pkg/amarra/live -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/amarra/live
git commit -m "feat(live): stub /amarra/live with 501 until slice B"
```

---

### Task 7: `amarra.js` Drive, Frame, Stream, Hook

**Files:**

- Create: `pkg/amarra/js/morph.mjs`, `drive.mjs`, `frame.mjs`, `stream.mjs`, `hook.mjs`, `entry.mjs`
- Test: `pkg/amarra/js/*.test.mjs`
- Modify: `package.json` `js:test` glob to `pkg/amarra/js/**/*.test.mjs`
- Modify: `scripts/js-build.mjs` to emit `pkg/cais/pwa/assets/amarra.js`

- [ ] **Step 1: Write the failing JS tests**

`pkg/amarra/js/drive.test.mjs`:

```js
import test from "node:test";
import assert from "node:assert/strict";
import { shouldInterceptClick, driveHeaders } from "./drive.mjs";

test("shouldInterceptClick ignores new tab and download", () => {
  assert.equal(
    shouldInterceptClick({
      href: "/x",
      target: "_blank",
      download: false,
      origin: "http://a",
      locationOrigin: "http://a",
    }),
    false
  );
  assert.equal(
    shouldInterceptClick({
      href: "/x",
      target: "",
      download: true,
      origin: "http://a",
      locationOrigin: "http://a",
    }),
    false
  );
});

test("shouldInterceptClick allows same-origin GET link", () => {
  assert.equal(
    shouldInterceptClick({
      href: "http://a/x",
      target: "",
      download: false,
      origin: "http://a",
      locationOrigin: "http://a",
    }),
    true
  );
});

test("driveHeaders sets Amarra-Drive and CSRF", () => {
  const h = driveHeaders("tok");
  assert.equal(h["Amarra-Drive"], "true");
  assert.equal(h["X-CSRF-Token"], "tok");
  assert.equal(h["Accept"], "text/html");
});
```

`pkg/amarra/js/stream.test.mjs`: parse an SSE chunk `event: append\ndata: <p>x</p>\n\n` into `{kind:"append", html:"<p>x</p>"}`.

`pkg/amarra/js/hook.test.mjs`: `csrfTokenFromMeta('<meta name="csrf-token" content="abc">') === "abc"`.

- [ ] **Step 2: Run tests to verify they fail**

```bash
node --test pkg/amarra/js/**/*.test.mjs
```

Expected: FAIL — modules missing.

- [ ] **Step 3: Implement modules**

`drive.mjs` (pure functions + `start(document, fetchFn, morphFn)`):

- Click: `a[href]` same origin, no `download`, no `target`, no `data-amarra-skip`.
- Submit: `form` without `data-amarra-skip`; POST `FormData`.
- `fetch` with `driveHeaders`, `redirect: "follow"`.
- On 422/200: `morph(document.querySelector("#amarra-main"), html)` by parsing `#amarra-main` from the response.
- On 401/403 CSRF: `location.reload()`.
- `history.pushState` on success GET and POST-after-redirect (use `response.url`).

`frame.mjs`: custom element `amarra-frame` with `id` + `src`; fetch with `Amarra-Frame: id`; replace innerHTML/morph inside the element.

`stream.mjs`: `EventSource` or fetch SSE; apply `append`/`prepend`/`replace`/`morph`/`remove`/`toast` (toast dispatches `amarra:toast`).

`hook.mjs`: listen `amarra:toast` → `#amarra-toast-host`; `data-amarra-focus`; optimistic `data-amarra-optimistic` (port the toggle/count/remove logic from `pkg/cais/pwa/assets/cais-core.js`); add CSRF on Drive requests via `driveHeaders`.

`morph.mjs`: `export function morph(el, html)` using vendored Idiomorph. Vendor file: copy Idiomorph ESM into `pkg/amarra/js/vendor/idiomorph.js` (do not reimplement). Tests pass a fake `morphFn`.

`entry.mjs`: `hook.start(); drive.start(); frame.define(); stream.start();` and `window.amarra = { drive, live: { connect() { /* no-op slice A */ } } }`.

`scripts/js-build.mjs`: esbuild `pkg/amarra/js/entry.mjs` → IIFE `pkg/cais/pwa/assets/amarra.js`.

- [ ] **Step 4: Run JS tests + build**

```bash
npm run js:test
npm run js:build
test -f pkg/cais/pwa/assets/amarra.js
```

Expected: PASS, file exists.

- [ ] **Step 5: Commit**

```bash
git add pkg/amarra/js pkg/cais/pwa/assets/amarra.js scripts/js-build.mjs package.json
git commit -m "feat(js): Amarra Drive, Frame, Stream, and Hook runtime"
```

---

### Task 8: `InstallForAmarra` PWA

**Files:**

- Modify: `pkg/cais/pwa/pwa.go`
- Test: `pkg/cais/pwa/pwa_test.go`
- Modify: SW template so it network-first caches HTML + `/static/js/amarra.js` (not `/static/build/`)

- [ ] **Step 1: Write the failing test**

```go
func TestInstallForAmarra_writesAmarraJS(t *testing.T) {
	dir := t.TempDir()
	if err := InstallForAmarra(dir, "Demo"); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{
		"web/static/js/amarra.js",
		"web/static/js/sw.js",
		"web/static/offline.html",
		"web/static/manifest.webmanifest",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "web/static/js/htmx.min.js")); err == nil {
		t.Error("htmx should not be installed")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/cais/pwa -run TestInstallForAmarra_writesAmarraJS -count=1
```

Expected: FAIL.

- [ ] **Step 3: Implement `InstallForAmarra`**

Copy `WriteStaticInertia` structure: manifest, icons, offline.html, og.png, `SyncServiceWorker`. Asset list:

- `assets/amarra.js` → `js/amarra.js`
- `assets/offline.html` → `offline.html`
- `assets/icon.png` → `icons/icon.png`

Do **not** copy htmx, sse-ext, idiomorph-ext, cais.js, cais-core.js.

SW: network-first for navigations and `/static/js/amarra.js` and `/static/css/`; fallback `offline.html`. Keep `CACHE_VERSION` bump behavior.

- [ ] **Step 4: Run tests**

```bash
go test ./pkg/cais/pwa -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add pkg/cais/pwa
git commit -m "feat(pwa): InstallForAmarra ships amarra.js without HTMX"
```

---

### Task 9: `amarra-cais new` HTML scaffold

**Files:**

- Modify: `internal/cli/scaffold.go` file lists
- Modify: `internal/cli/tpl_scaffold_web.go` layout (`#amarra-main`, one script `/static/js/amarra.js`)
- Replace Svelte page templates with HTML pages under `web/templates/pages/`
- Replace Inertia handlers with `view.Write`
- Modify: `tplGoMod` (drop gonertia, require `github.com/puppe1990/amarra-cais`)
- Delete from scaffold: `vite.config.js`, `svelte.config.js`, `web/src/**`, `inertia_test.go`
- Test: `internal/cli/cli_new_test.go`, `internal/cli/scaffold_templates_test.go`

- [ ] **Step 1: Write the failing scaffold test**

```go
func TestScaffoldNewApp_htmlFirstNoInertia(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	appDir := filepath.Join(t.TempDir(), "demo")
	if err := scaffoldNewApp(appDir, scaffoldData{AppName: "demo", ModulePath: "example.com/demo", CaisVersion: "0.1.0"}, false, false); err != nil {
		t.Fatal(err)
	}
	mustExist := []string{
		"web/templates/layouts/app.html",
		"web/templates/pages/home.html",
		"web/templates/pages/login.html",
		"web/templates/components/.gitkeep",
		"web/static/js/amarra.js",
		"internal/handlers/home.go",
	}
	mustNot := []string{
		"vite.config.js",
		"svelte.config.js",
		"web/src/pages/Home.svelte",
		"web/src/main.js",
		"internal/handlers/inertia_test.go",
	}
	for _, p := range mustExist {
		if _, err := os.Stat(filepath.Join(appDir, p)); err != nil {
			t.Errorf("missing %s", p)
		}
	}
	for _, p := range mustNot {
		if _, err := os.Stat(filepath.Join(appDir, p)); err == nil {
			t.Errorf("should not exist %s", p)
		}
	}
	gomod, _ := os.ReadFile(filepath.Join(appDir, "go.mod"))
	if strings.Contains(string(gomod), "gonertia") {
		t.Error("go.mod still has gonertia")
	}
	if !strings.Contains(string(gomod), "github.com/puppe1990/amarra-cais") {
		t.Error("go.mod missing amarra-cais")
	}
	home, _ := os.ReadFile(filepath.Join(appDir, "internal/handlers/home.go"))
	if strings.Contains(string(home), "inertia") {
		t.Error("home handler still uses inertia")
	}
	if !strings.Contains(string(home), "view.Write") {
		t.Error("home handler missing view.Write")
	}
}
```

Update `TestScaffoldNewApp_includesAgentsMD` needles: `Amarra`, `view.Write`, `amarra-cais g`, `web/templates/pages` — **not** `Inertia` or `web/src/pages`.

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/cli -run TestScaffoldNewApp_htmlFirstNoInertia -count=1
```

Expected: FAIL.

- [ ] **Step 3: Rewrite templates**

Layout (`web/templates/layouts/app.html` in the app):

- `<meta name="csrf-token">` as today
- `<script src="/static/js/amarra.js" defer></script>` only
- `<body>` without `hx-ext`
- `<nav id="amarra-nav">`, `<main id="amarra-main">`, `<div id="amarra-toast-host">`
- `<.flash />` after expand, or `{{ template }}` flash partial if kit flash needs `.Flash`

Home handler:

```go
func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	view.Write(w, r, h.views, view.Page{
		Layout: "app",
		Name:   "home",
		Data: map[string]any{
			"Title":     h.catalog.T("home.title"),
			"Site":      meta.ForRequest(h.site, r),
			"CSRFToken": csrf.FromRequest(r), // use the real Cais CSRF helper already used in templates
			"Flash":     flash.MessageFromRequest(r),
		},
	}, h.cfg)
}
```

Constructor drops `*inertia.Inertia`. `internal/app` deps struct drops Inertia; adds `*view.Renderer` loaded from `web/templates`. Register `live.Handler` at `/amarra/live`.

`tplGoMod`:

```
require (
	github.com/puppe1990/amarra-cais v{{.CaisVersion}}
	modernc.org/sqlite v1.53.0
)
```

`package.json` of the app: Tailwind only (keep current Tailwind scripts; remove Vite/Svelte/Inertia deps).

`cais dev` later (Task 11) must not start Vite; scaffold `.air.toml` stays Go-only.

Call `pwa.InstallForAmarra` instead of `InstallForInertia`.

Home/contact/login HTML pages: port copy from the Svelte pages to `{{ define "content" }}` using `<.form>` / `<.input>` / `<.button>`. Keep i18n `{{ t "key" }}`.

Handler tests: `httptest` + `strings.Contains` HTML; delete `setupTestInertia`.

- [ ] **Step 4: Run CLI tests**

```bash
go test ./internal/cli -count=1
```

Expected: PASS. Fix every template test that still looks for Svelte/Inertia.

- [ ] **Step 5: Commit**

```bash
git add internal/cli pkg/cais/pwa
git commit -m "feat(cli): scaffold HTML Amarra apps without Inertia"
```

---

### Task 10: Generators (handler, resource, auth, component, destroy)

**Files:**

- Modify: `internal/cli/resource_gen_inertia.go` → HTML/Amarra (or replace usage with `resource_gen_html.go` and delete inertia gen)
- Modify: `internal/cli/scaffold_handler.go`, `scaffold_auth.go`, `destroy.go`
- Create: `internal/cli/scaffold_component.go`
- Test: existing `resource_scaffold_*_test.go`, `scaffold_handler_test.go`, `scaffold_auth_test.go`, new `scaffold_component_test.go`

- [ ] **Step 1: Write failing tests**

```go
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

func TestGenerateHandler_writesHTMLNotSvelte(t *testing.T) {
	t.Setenv("CAIS_SKIP_TIDY", "1")
	dir := filepath.Join(t.TempDir(), "app")
	if err := scaffoldNewApp(dir, scaffoldData{AppName: "app", ModulePath: "example.com/app"}, true, false); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	if err := (&CLI{Out: io.Discard}).Run([]string{"g", "handler", "settings"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web/templates/pages/settings.html")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web/src/pages/Settings.svelte")); err == nil {
		t.Fatal("svelte page should not exist")
	}
	src, _ := os.ReadFile(filepath.Join(dir, "internal/handlers/settings.go"))
	if strings.Contains(string(src), "inertia") {
		t.Fatal("handler still references inertia")
	}
	if !strings.Contains(string(src), "view.Write") {
		t.Fatal("handler missing view.Write")
	}
}
```

Resource test: generated admin form contains `<.form` or expanded `<form` + `csrfField`; no `AdminItems.svelte`.

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/cli -run 'TestGenerateComponent_|TestGenerateHandler_writesHTMLNotSvelte' -count=1
```

Expected: FAIL.

- [ ] **Step 3: Implement generators**

- `g handler name` → `internal/handlers/name.go`, `name_test.go`, `web/templates/pages/name.html`, AST route patch. No Svelte.
- `g resource` → always HTML path (`resource_gen_html.go`). Swap `hxForm` / `hxPaginate` for `<.form>` + `{{ linkTo }}` + `<amarra-frame id="admin-items">` wrapping the table. Delete inertia resource output.
- `g auth` → HTML pages already in `new`; keep idempotent HTML path.
- `g component name` → `web/templates/components/name.html` with `{{ .Inner }}` stub + `components/name_test.go` in framework? App-side: a `internal/handlers` is wrong. Only the HTML file + mention in dry-run. Framework tests render it via `view.Load` on the app dir if cheap; otherwise just file existence.
- `g stream chat` → templates without hx-ext; use `data-amarra` + Stream SSE (`WriteOp`). Keep Go chat handlers, retarget HTML ids.
- `destroy` → delete `.html` pages/components; stop deleting `.svelte`.

`cli.go` help: add `amarra-cais g component <name>`.

- [ ] **Step 4: Run CLI tests**

```bash
go test ./internal/cli -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/cli
git commit -m "feat(cli): generate Amarra HTML handlers, resources, and components"
```

---

### Task 11: `dev` / `install` / `build` / `doctor` without Vite

**Files:**

- Modify: `internal/cli/frontend.go`, `commands.go`, `doctor.go`, `doctor_test.go`, `doctor_mobile.go`
- Test: `internal/cli/frontend_test.go`, `internal/cli/doctor_test.go`

- [ ] **Step 1: Write failing tests**

```go
func TestHasViteApp_falseOnHtmlScaffold(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"scripts":{"css":"tailwindcss"}}`), 0o644)
	if hasViteApp(dir) {
		t.Fatal("html app must not look like Vite")
	}
}

func TestDoctor_requiresAmarraJS(t *testing.T) {
	// scaffold minimal with CAIS_SKIP_TIDY
	// runDoctor
	// output contains "amarra.js" ok
	// does not contain "Inertia frontend" failure
}
```

Doctor must **fail** if `vite.config.js` exists on a new app (Inertia leftover) and **fail** if `web/static/js/amarra.js` missing.

- [ ] **Step 2: Run tests to verify they fail**

```bash
go test ./internal/cli -run 'TestHasViteApp_falseOnHtmlScaffold|TestDoctor_requiresAmarraJS' -count=1
```

Expected: FAIL (doctor still wants Inertia or HTMX).

- [ ] **Step 3: Implement**

- `hasViteApp` remains for _legacy_ detection only; `startViteWatch` already no-ops when false. `cmdDev` must not require `node_modules` for Vite; still run Tailwind watch if `input.css` exists.
- `cmdInstall`: `go mod tidy`; `npm install` only if `package.json` exists; **do not** `npm run build` Vite.
- `cmdBuild`: skip `runViteBuild`; require `web/static/css/styles.css` (existing CSS check).
- `runDoctor`: remove Inertia/Vite/HTMX/sse-ext as the default path. Add `checkAmarraJS` (`web/static/js/amarra.js`) and `checkAmarraLayout` (`#amarra-main` in `layouts/app.html`). If `vite.config.js` exists: fail with hint “this is Cais v0.11 Inertia; amarra-cais does not support it”.
- Help text: `dev` = air + tailwind, not vite.
- Mobile doctor: drop checks that require `cais.js` / htmx sse-ext; check `amarra.js` presence instead.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/cli -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/cli
git commit -m "feat(cli): drop Vite from dev, build, install, and doctor"
```

---

### Task 12: Smoke, README, AGENTS, framework docs

**Files:**

- Modify: `scripts/smoke-scaffold.sh`, `scripts/smoke-production.sh`
- Modify: `README.md`, `AGENTS.md`, generated `tpl_scaffold_readme.go`, `tpl_scaffold_agents.go`
- Modify: `package.json` if js:test glob not already updated

- [ ] **Step 1: Write the failing smoke expectation**

Update `scripts/smoke-scaffold.sh`:

```bash
go build -o "$TMP/amarra-cais" ./cmd/amarra-cais
"$TMP/amarra-cais" new smokeapp "$APP"
cd "$APP"
go mod tidy
go test ./... -count=1
go build -o "$TMP/server" ./cmd/server
grep -q 'id="amarra-main"' web/templates/layouts/app.html
test -f web/static/js/amarra.js
! test -f vite.config.js
```

- [ ] **Step 2: Run smoke to verify it fails** (if scaffold still incomplete, fix; this step is the integration gate)

```bash
bash scripts/smoke-scaffold.sh
```

Expected: PASS after Task 9–11. If FAIL, fix the pointed file; do not weaken the grep.

- [ ] **Step 3: Docs**

README stack table: Amarra Views + Drive (not Inertia + Svelte). Quick start: `go install github.com/puppe1990/amarra-cais/cmd/amarra-cais@v0.1.0`. Commands table uses `amarra-cais`.

Framework `AGENTS.md`: handlers take `*view.Renderer`, not `*inertia.Inertia`; pages are `web/templates/pages/*.html`; TDD is `go test` + `npm run js:test`.

Generated app AGENTS: same.

- [ ] **Step 4: Full CI locally**

```bash
make test
npm run js:test
bash scripts/smoke-scaffold.sh
```

Expected: all PASS. `make ci` if lint/format hooks still apply (run `gofmt` / prettier on touched files).

- [ ] **Step 5: Commit**

```bash
git add README.md AGENTS.md scripts internal/cli/tpl_scaffold_readme.go internal/cli/tpl_scaffold_agents.go
git commit -m "docs: amarra-cais HTML-first README, AGENTS, and smoke"
```

Tag when publishing: `v0.1.0`. Do not tag until smoke is green on a clean `amarra-cais new`.

---

## Self-review (plan vs spec)

| Spec section                          | Task                                     |
| ------------------------------------- | ---------------------------------------- |
| Identity / module / CLI `amarra-cais` | 1                                        |
| Amarra Views preprocessor + kit       | 2, 4                                     |
| `view.Write` Drive/Frame              | 3                                        |
| Stream named ops                      | 5                                        |
| Live stub 501                         | 6                                        |
| `amarra.js` Drive/Frame/Stream/Hook   | 7                                        |
| PWA `InstallForAmarra`                | 8                                        |
| `new` HTML scaffold, no Inertia       | 9                                        |
| generators + destroy                  | 10                                       |
| dev/install/build/doctor              | 11                                       |
| smoke + docs                          | 12                                       |
| 422 / CSRF / flash                    | 9 (handlers + kit); Drive 422 morph in 7 |
| Slice B Live hub                      | **not this plan**                        |
| jobsui rewrite                        | **not this plan**                        |
| `{@x}` / templ / named slots          | out of spec                              |

No TBD in tasks. Types: `view.Page`, `HeaderDrive`, `HeaderFrame`, `stream.Op`, `live.Handler` are consistent across tasks.
