# Sidebar fixa pós-login Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Scaffold `layouts/app.html` nasce com sidebar fixa à esquerda (Dashboard + logout, `<!-- cais:nav -->` dentro) em vez de nav horizontal, nas 3 variações.

**Architecture:** Só templates em `internal/cli/tpl_scaffold_web.go` + ajustes nos testes de layout. O `id="amarra-nav"` muda de `<nav>` para `<aside>` (hook do Drive, marker de patch e testes de shell continuam valendo). Drawer mobile sem JS via checkbox + `peer-checked`.

**Tech Stack:** Go templates (consts `tpl*`), Tailwind (`peer-checked`, `lg:`), `go test ./internal/cli/`.

---

### Task 1: Teste failing do shell sidebar

**Files:**
- Modify: `internal/cli/tpl_scaffold_layout_test.go`

- [ ] **Step 1: Adicionar o teste**

Acrescente ao final de `internal/cli/tpl_scaffold_layout_test.go`:

```go
func TestLayoutTemplates_sidebarShell(t *testing.T) {
	for name, tpl := range map[string]string{
		"full":    tplLayout,
		"minimal": tplLayoutMinimal,
		"blank":   tplLayoutBlank,
	} {
		for _, token := range []string{
			`<aside id="amarra-nav"`,
			`fixed left-0 top-[57px] bottom-0`,
			`w-60`,
			`lg:ml-60`,
			`amarra-sidebar-toggle`,
			`peer-checked:translate-x-0`,
			`amarra-hook="nav"`,
			`href="/dashboard"`,
			`action="/logout"`,
			`<!-- cais:nav -->`,
		} {
			if !strings.Contains(tpl, token) {
				t.Errorf("%s layout missing sidebar token %q", name, token)
			}
		}
		for _, gone := range []string{
			`<nav id="amarra-nav"`,
			`template "nav_links"`,
		} {
			if strings.Contains(tpl, gone) {
				t.Errorf("%s layout should not contain %q", name, gone)
			}
		}
	}
}
```

- [ ] **Step 2: Rodar e ver falhar**

Run: `go test ./internal/cli/ -run TestLayoutTemplates_sidebarShell -count=1 -v`
Expected: FAIL (templates ainda têm `<nav id="amarra-nav"` e `template "nav_links"`).

- [ ] **Step 3: Commit do teste (red)**

```bash
git add internal/cli/tpl_scaffold_layout_test.go
git commit -m "test(cli): sidebar shell tokens for scaffold layouts"
```

---

### Task 2: Shell sidebar no template

**Files:**
- Modify: `internal/cli/tpl_scaffold_web.go:38-72`

- [ ] **Step 1: Trocar header/nav por sidebar**

Em `tplLayoutBaseOpen`, logo após `<div>` (linha 39), insira o checkbox do drawer:

```html
<input id="amarra-sidebar-toggle" type="checkbox" class="peer sr-only" aria-label="Menu" />
```

Na linha do header (dentro do `div.max-w-6xl...`, antes do logo), insira o hamburger visível só no mobile:

```html
<label for="amarra-sidebar-toggle" class="lg:hidden font-mono text-[10px] uppercase tracking-[0.22em] text-copper cursor-pointer">Menu</label>
```

Substitua o bloco (linhas 54-58):

```html
<nav id="amarra-nav" class="bg-ink border-b border-foam/10 sticky top-[57px] z-30">
  <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
    <!-- nav hook re-syncs active link after Drive morph (#27); SSR ActiveNav stays the first-paint default -->
    <div amarra-hook="nav" data-amarra-nav-on="text-copper" data-amarra-nav-off="text-foam/50 hover:text-foam" class="flex gap-1 py-1 overflow-x-auto no-scrollbar">
```

por:

```html
<aside id="amarra-nav" class="bg-ink border-r border-foam/10 fixed left-0 top-[57px] bottom-0 z-30 w-60 -translate-x-full transition-transform peer-checked:translate-x-0 lg:translate-x-0">
  <div class="h-full flex flex-col gap-1 p-3 overflow-y-auto no-scrollbar">
    <!-- nav hook re-syncs active link after Drive morph (#27); SSR ActiveNav stays the first-paint default -->
    <div amarra-hook="nav" data-amarra-nav-on="text-copper" data-amarra-nav-off="text-foam/50 hover:text-foam" class="flex flex-col gap-1">
```

Em `tplLayoutBaseClose`, o fechamento antigo de 3 níveis (`</div>` do hook + `</div>` do max-w-6xl + `</nav>`) vira 3 níveis novos (`</div>` do hook + `</div>` do h-full + `</aside>`). E adicione `lg:ml-60` ao `<main id="amarra-main">`:

```html
<main id="amarra-main" class="flex-grow px-4 sm:px-6 lg:px-8 py-5 lg:ml-60">
```

Atenção: o checkbox `peer` precisa ser irmão anterior do `aside` (mesmo pai). O `<div>` da linha 39 é pai de header, aside e main — ok.

- [ ] **Step 2: Rodar teste do shell (ainda pode falhar por nav_links — esperado)**

Run: `go test ./internal/cli/ -run 'TestLayoutTemplates_sidebarShell|TestLayoutTemplates_hasDriveShell|TestLayoutTemplates_shellDesignTokens' -count=1 -v`
Expected: sidebarShell ainda FAIL em `template "nav_links"` (Task 3 remove); hasDriveShell e shellDesignTokens PASS (`id="amarra-nav"`, `sticky top-0` do header e `no-scrollbar` preservados).

- [ ] **Step 3: Commit parcial**

```bash
git add internal/cli/tpl_scaffold_web.go
git commit -m "feat(cli): scaffold shell with fixed sidebar instead of top nav"
```

---

### Task 3: Itens default + remover nav_links

**Files:**
- Modify: `internal/cli/tpl_scaffold_web.go:60-64,103-105`
- Modify: `internal/cli/tpl_scaffold_partials.go:33-37` (deletar `tplPartialNavLinks`)
- Modify: `internal/cli/tpl_scaffold_layout_test.go` (atualizar testes afetados)

- [ ] **Step 1: Sidebar nasce com Dashboard + logout**

Substitua:

```go
const tplLayoutNavFull = `<!-- cais:nav -->
            {{"{{"}} template "nav_links" . {{"}}"}}
            <.locale-toggle current="{{"{{"}} .Locale {{"}}"}}" />`

const tplLayoutNavEmpty = `<!-- cais:nav -->`
```

por (mesmo conteúdo nos dois — full, minimal e blank nascem idênticos na sidebar):

```go
const tplLayoutNavFull = `<a href="/dashboard" class="px-3 py-2 font-mono text-[10px] uppercase tracking-[0.22em] transition flex items-center gap-2 flex-shrink-0 {{"{{"}} if eq .ActiveNav "dashboard" {{"}}"}}text-copper{{"{{"}} else {{"}}"}}text-foam/50 hover:text-foam{{"{{"}} end {{"}}"}}">{{"{{"}} template "icon_chart_nav" . {{"}}"}}Dashboard</a>
            <!-- cais:nav -->
            <.form action="/logout" method="post">
              <.button type="submit">Sair</.button>
            </.form>`

const tplLayoutNavEmpty = tplLayoutNavFull
```

Remova `tplPartialNavLinks` de `tpl_scaffold_partials.go` e das concatenações em `tplLayout`/`tplLayoutMinimal` (linhas 103-105). `icon_chart_nav` continua vindo de `tplPartialIcons` — não mexer.

- [ ] **Step 2: Atualizar testes que citavam nav_links**

Em `internal/cli/tpl_scaffold_layout_test.go`:
- Deletar `TestLayoutTemplates_fullHasDefaultNavLinks` (trocado pelo sidebarShell da Task 1; toast-host segue coberto? mover a asserção `amarra-toast-host` para o sidebarShell não — ela já existe só nesse teste; adicione `amarra-toast-host` à lista de tokens do `TestLayoutTemplates_sidebarShell`).
- Deletar `TestLayoutTemplates_navTabsHaveIcons` (parcial removida). O ícone `icon_chart_nav` segue coberto por `TestScaffoldPartials_iconsRenderNonEmpty`? Não — ele testa só icons.html. Adicione asserção em sidebarShell: token `template "icon_chart_nav"`.
- Em `TestScaffoldPartials_iconsRenderNonEmpty`: remover a entrada `nav_links.html` do map e o branch `want` correspondente.
- Em `TestScaffoldPages_dropIndigoAndHTMX`: remover a entrada `"nav": tplPartialNavLinks` do map `blobs`.

Adicione os dois tokens (`amarra-toast-host`, `template "icon_chart_nav"`) ao teste da Task 1 via edit antes de rodar.

- [ ] **Step 3: Rodar suite do CLI**

Run: `go test ./internal/cli/ -count=1 2>&1 | tail -n 5`
Expected: ok (sem FAIL).

- [ ] **Step 4: Commit**

```bash
git add internal/cli/tpl_scaffold_web.go internal/cli/tpl_scaffold_partials.go internal/cli/tpl_scaffold_layout_test.go
git commit -m "feat(cli): sidebar defaults to Dashboard+logout, drop nav_links partial"
```

---

### Task 4: Validação ponta a ponta + push

- [ ] **Step 1: Smoke — scaffold gera e compila**

```bash
rm -rf /var/folders/8d/b6k_9tbx4zd77vkck6rdy9mh0000gn/T/opencode/sbtest && go run ./cmd/amarra-cais new sbtest --minimal 2>&1 | tail -n 3
go build ./... 2>&1 | tail -n 3
```

Run (workdir): `/var/folders/8d/b6k_9tbx4zd77vkck6rdy9mh0000gn/T/opencode` para o `new`; `go build` dentro de `sbtest`.
Expected: scaffold ok; build ok sem erros. Conferir `sbtest/web/templates/layouts/app.html` contém `<aside id="amarra-nav"`.

- [ ] **Step 2: Gate completo**

Run: `go test ./... -count=1 2>&1 | grep -E "FAIL" ; make lint 2>&1 | tail -n 3; gofmt -l internal/cli/`
Expected: nenhum FAIL, `0 issues`, gofmt vazio.

- [ ] **Step 3: Push em branch + PR**

```bash
git checkout -b feat-sidebar-shell && git push -u origin feat-sidebar-shell
gh pr create --title "scaffold sidebar fixa pós-login" --body "Troca nav horizontal por sidebar fixa (Dashboard+logout, marker cais:nav dentro). Spec: docs/superpowers/specs/2026-09-16-sidebar-shell-design.md"
```

Não mergear sem CI verde e sem aprovação do usuário.
