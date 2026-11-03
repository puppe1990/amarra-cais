package cli

// Amarra layout for cais new (full, minimal, blank). One shell: #amarra-main + amarra.js.
const tplLayoutTitleDesc = `{{"{{"}} define "title" {{"}}"}}{{.AppName}}{{"{{"}} end {{"}}"}}
{{"{{"}} define "description" {{"}}"}}{{.AppName}} — powered by Amarra{{"{{"}} end {{"}}"}}`

const tplLayoutBrand = `<a href="/" class="group flex items-center gap-3">
            <span class="flex h-9 w-9 items-center justify-center rounded-lg border border-foam/15 text-copper transition-colors group-hover:border-copper/60" aria-hidden="true">
              <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5 20V8m0 0a3 3 0 116 0v1H5m14 11V9a3 3 0 00-3-3h-3" /></svg>
            </span>
            <span class="min-w-0">
              <span class="block truncate font-serif text-lg font-medium leading-tight text-foam">{{.AppName}}</span>
              <span class="mt-1 block font-mono text-[11px] leading-none text-foam/50" title="Cais v{{"{{"}} .FrameworkVersionFull {{"}}"}}">Amarra v{{"{{"}} .FrameworkVersion {{"}}"}}</span>
            </span>
          </a>`

const tplLayoutMenuButton = `<button type="button" id="amarra-sidebar-toggle" class="inline-flex h-11 w-11 items-center justify-center rounded-lg border border-foam/15 text-foam/70 hover:bg-foam/5 hover:text-foam lg:hidden" aria-label="Menu" aria-controls="amarra-nav" aria-expanded="false" amarra-hook="sidebar" data-amarra-sidebar-target="#amarra-nav">
            <svg class="h-[18px] w-[18px]" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M4 7h16M4 12h16M4 17h16" /></svg>
          </button>`

// Shared by the anonymous header and the signed-in sidebar so the theme
// control never drifts between shells (#258). The tile shows the theme the
// button turns on — sun while the app is dark, moon while it is light; the swap
// is CSS-only on html.light, so the hook keeps owning just the label slot. The
// label collapses to the icon on small screens so the header never wraps.
const tplLayoutThemeToggle = `<button type="button" class="group inline-flex h-8 items-center gap-2 rounded-lg px-2 text-foam/70 hover:bg-foam/5 hover:text-foam" amarra-hook="theme" aria-pressed="false" data-amarra-theme-color="#f7f1e6" data-amarra-theme-color-off="#081318" data-amarra-theme-on-label="{{"{{"}} t "layout.theme_dark" {{"}}"}}" data-amarra-theme-off-label="{{"{{"}} t "layout.theme_light" {{"}}"}}">
            <span class="flex h-4 w-4 flex-shrink-0 items-center justify-center" aria-hidden="true">
              <span class="[html.light_&]:hidden">{{"{{"}} template "icon_sun" . {{"}}"}}</span>
              <span class="hidden [html.light_&]:block">{{"{{"}} template "icon_moon" . {{"}}"}}</span>
            </span>
            <span class="hidden whitespace-nowrap text-sm sm:inline" data-amarra-theme-label>{{"{{"}} t "layout.theme_light" {{"}}"}}</span>
          </button>`

// tplLayoutAuthLink goes into the anonymous header of the full scaffold only:
// minimal/blank ship no /login route, so the link would 404 on a fresh app (#336).
const tplLayoutAuthLink = `<a href="/login" class="inline-flex h-8 items-center whitespace-nowrap rounded-lg px-3 text-sm text-foam/70 hover:bg-foam/5 hover:text-foam">{{"{{"}} t "auth.login_title" {{"}}"}}</a>`

const tplLayoutBaseOpenPre = `{{"{{"}} define "app" {{"}}"}}
<!doctype html>
<html lang="{{"{{"}} htmlLang {{"}}"}}" data-amarra-layout="app">
  <head>
    <meta charset="UTF-8" />
    <script nonce="{{"{{"}} .CSPNonce {{"}}"}}">
      try {
        if (localStorage.getItem("amarra-theme") === "light") document.documentElement.classList.add("light");
      } catch (e) {}
    </script>
    <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover" />
    {{"{{"}} if .CSRFToken {{"}}"}}<meta name="csrf-token" content="{{"{{"}} .CSRFToken {{"}}"}}" />{{"{{"}} end {{"}}"}}
    <title>{{"{{"}} if .Title {{"}}"}}{{"{{"}} .Title {{"}}"}} · {{.AppName}}{{"{{"}} else {{"}}"}}{{.AppName}}{{"{{"}} end {{"}}"}}</title>
    <meta name="description" content="{{.AppName}} — powered by Amarra" />
    <meta property="og:type" content="website" />
    <meta property="og:site_name" content="{{.AppName}}" />
    <meta property="og:title" content="{{"{{"}} if .Title {{"}}"}}{{"{{"}} .Title {{"}}"}}{{"{{"}} else {{"}}"}}{{.AppName}}{{"{{"}} end {{"}}"}}" />
    <meta property="og:image" content="{{"{{"}} absURL .Site.AppURL "/static/og.png" {{"}}"}}" />
    <meta property="og:locale" content="{{"{{"}} ogLocale {{"}}"}}" />
    <meta name="twitter:card" content="summary_large_image" />
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
    <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Fraunces:opsz,wght@9..144,500;9..144,620&display=swap" />
    <link rel="stylesheet" href="/static/css/styles.css" />
    <link rel="manifest" href="/static/manifest.webmanifest" />
    <meta name="theme-color" content="#081318" />
    <meta name="mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
    <meta name="apple-mobile-web-app-title" content="{{.AppName}}" />
    <link rel="apple-touch-icon" href="/static/icons/icon.png" />
    <link rel="icon" href="/static/favicon.svg" type="image/svg+xml" />
    <script src="/static/js/amarra.js" defer></script>
  </head>
  <body class="min-h-screen bg-ink font-sans antialiased text-foam flex flex-col justify-between" data-amarra-shell="{{"{{"}} if .Site.LoggedIn {{"}}"}}app{{"{{"}} else {{"}}"}}auth{{"{{"}} end {{"}}"}}">
    <div>
      {{"{{"}} if .Site.LoggedIn {{"}}"}}
      <!-- app shell: desktop is the fixed rail; mobile keeps a slim brand + menu bar -->
      <header class="lg:hidden bg-ink border-b border-foam/10 sticky top-0 z-40">
        <div class="px-4 h-14 flex items-center gap-4">
          ` + tplLayoutMenuButton + `
          ` + tplLayoutBrand + `
        </div>
      </header>
      {{"{{"}} end {{"}}"}}
      {{"{{"}} if not .Site.LoggedIn {{"}}"}}
      <!-- auth shell: no sidebar; language, theme and login live in the header -->
      <header class="bg-ink border-b border-foam/10 sticky top-0 z-40">
        <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 h-14 flex items-center justify-between gap-4">
          ` + tplLayoutBrand + `
          <div class="flex items-center gap-1">
            <.locale-toggle current="{{"{{"}} .Locale {{"}}"}}" />
            ` + tplLayoutThemeToggle + `
            `

const tplLayoutBaseOpenPost = `
          </div>
        </div>
      </header>
      {{"{{"}} end {{"}}"}}
      {{"{{"}} if .Site.LoggedIn {{"}}"}}
      <aside id="amarra-nav" class="bg-ink border-r border-foam/10 fixed left-0 top-14 bottom-0 z-30 w-60 -translate-x-full invisible transition-transform data-[amarra-sidebar-open]:translate-x-0 data-[amarra-sidebar-open]:visible lg:translate-x-0 lg:visible lg:top-0">
        <div class="h-full flex flex-col gap-1 p-4 overflow-y-auto no-scrollbar">
          <div class="hidden lg:flex items-center pb-4 mb-3 border-b border-foam/10">` + tplLayoutBrand + `</div>
          <!-- nav hook re-syncs active link after Drive morph (#27); SSR ActiveNav stays the first-paint default -->
          <div amarra-hook="nav" data-amarra-nav-on="bg-foam/5 text-foam" data-amarra-nav-off="text-foam/60 hover:bg-foam/5 hover:text-foam" class="flex flex-col gap-1">
            `

// tplLayoutBaseOpen keeps the /login link for the full scaffold;
// tplLayoutBaseOpenSlim drops it for minimal/blank, which ship no auth routes (#336).
const tplLayoutBaseOpen = tplLayoutBaseOpenPre + tplLayoutAuthLink + tplLayoutBaseOpenPost

const tplLayoutBaseOpenSlim = tplLayoutBaseOpenPre + tplLayoutBaseOpenPost

const tplLayoutNavFull = `<a href="/dashboard" class="amarra-nav-link relative flex h-9 items-center rounded-lg pl-4 pr-3 text-sm flex-shrink-0 {{"{{"}} if eq .ActiveNav "dashboard" {{"}}"}}bg-foam/5 text-foam{{"{{"}} else {{"}}"}}text-foam/60 hover:bg-foam/5 hover:text-foam{{"{{"}} end {{"}}"}}"{{"{{"}} if eq .ActiveNav "dashboard" {{"}}"}} aria-current="page"{{"{{"}} end {{"}}"}}>Dashboard</a>
            <!-- cais:nav -->`

const tplLayoutNavEmpty = tplLayoutNavFull

const tplLayoutBaseClose = `
          </div>
          <div class="mt-auto flex flex-col gap-2 border-t border-foam/10 pt-4">
            <div class="flex items-center justify-between gap-2">
              <.locale-toggle current="{{"{{"}} .Locale {{"}}"}}" />
              ` + tplLayoutThemeToggle + `
            </div>
            <.form action="/logout" method="post">
              <button type="submit" class="flex h-9 w-full items-center rounded-lg pl-4 pr-3 text-left text-sm text-foam/60 hover:bg-foam/5 hover:text-foam">{{"{{"}} t "auth.logout" {{"}}"}}</button>
            </.form>
          </div>
        </div>
      </aside>
      {{"{{"}} end {{"}}"}}
      <div id="amarra-toast-host" aria-live="polite"></div>
      <main id="amarra-main" class="flex-grow px-4 py-6 sm:px-6 lg:px-8{{"{{"}} if .Site.LoggedIn {{"}}"}} lg:ml-60{{"{{"}} end {{"}}"}}">
        <.flash />
        {{"{{"}} template "content" . {{"}}"}}
      </main>
    </div>
    {{"{{"}} if not .Site.LoggedIn {{"}}"}}
    <footer class="mt-auto border-t border-foam/10 py-6">
      <p class="text-center text-xs text-foam/50">© 2026 {{.AppName}} · built with Amarra</p>
    </footer>
    {{"{{"}} end {{"}}"}}
    {{"{{"}} if eq .Site.Env "development" {{"}}"}}
    <script nonce="{{"{{"}} .CSPNonce {{"}}"}}">
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.getRegistrations().then(function (regs) {
          regs.forEach(function (r) { r.unregister(); });
        });
        if ("caches" in window) {
          caches.keys().then(function (keys) {
            keys.forEach(function (k) { caches.delete(k); });
          });
        }
      }
      // Dev reload (#340): the server re-parses templates in-process and pings
      // this stream, so an HTML/CSS edit refreshes the tab without an air
      // rebuild. Loopback only — the route is localhost-only and the phone on
      // the LAN URL would just collect 403s. The boot id flips when air
      // restarts the process for a .go edit; a changed id on reconnect means
      // the new binary is up, so reload once.
      (function () {
        if (!("EventSource" in window)) return;
        if (!/^(localhost|127\.0\.0\.1|\[::1\])$/.test(location.hostname)) return;
        var es = new EventSource("/_dev/reload");
        es.addEventListener("reload", function () { location.reload(); });
        es.addEventListener("boot", function (ev) {
          var seen = sessionStorage.getItem("amarra-dev-boot");
          sessionStorage.setItem("amarra-dev-boot", ev.data);
          if (seen && seen !== ev.data) location.reload();
        });
      })();
    </script>
    {{"{{"}} else {{"}}"}}
    <script nonce="{{"{{"}} .CSPNonce {{"}}"}}">
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/static/js/sw.js", { scope: "/" });
      }
    </script>
    {{"{{"}} end {{"}}"}}
  </body>
</html>
{{"{{"}} end {{"}}"}}`

const tplLayout = tplLayoutTitleDesc + tplPartialIcons + tplLayoutBaseOpen + tplLayoutNavFull + tplLayoutBaseClose

const tplLayoutMinimal = tplLayoutTitleDesc + tplPartialIcons + tplLayoutBaseOpenSlim + tplLayoutNavEmpty + tplLayoutBaseClose

const tplLayoutBlank = tplLayoutMinimal

const tplLayoutWelcome = tplLayoutMinimal

const tplCaisLogo = `{{"{{"}} define "cais_logo" {{"}}"}}
<img
  src="/static/img/go-on-cais.jpg"
  alt="Go on Cais"
  width="1024"
  height="683"
  class="w-full max-w-lg rounded-xl border border-foam/10"
/>
{{"{{"}} end {{"}}"}}
`

const tplPageHomePre = `{{"{{"}} define "content" {{"}}"}}
<section class="mx-auto flex max-w-6xl flex-col gap-12 lg:min-h-[calc(100svh-11rem)] lg:flex-row lg:items-center lg:gap-16">
  <div class="amarra-rise max-w-xl">
    <h1 class="font-serif text-[clamp(2.25rem,5vw,3.75rem)] font-medium leading-[1.05] text-foam">{{"{{"}} t "home.hero_heading" .Site.AppName {{"}}"}}</h1>
    <p data-testid="amarra-ready" class="mt-4 max-w-md text-lg leading-relaxed text-foam/70">{{"{{"}} t "home.tagline" {{"}}"}}</p>
    `

// The boarding CTA needs a /login route; minimal/blank ship none, so their home
// page reuses the same landing without it (#336). The label names where the flow
// lands (the dashboard) so it never doubles the header's "Sign in" link.
const tplPageHomeCTA = `<div class="mt-8">
      <a href="/login" class="amarra-btn amarra-btn-primary h-12 px-5 text-[15px]">{{"{{"}} t "home.cta_open" {{"}}"}}<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5 12h14m0 0l-6-6m6 6l-6 6" /></svg></a>
    </div>`

const tplPageHomePost = `
  </div>
  <aside class="amarra-rise amarra-rise-delay-1 w-full lg:max-w-[26rem]">
    <div class="overflow-hidden rounded-xl border border-foam/10 bg-tide/40">
      <div class="flex items-center justify-between gap-4 border-b border-foam/10 px-5 py-3">
        <p class="amarra-tick text-xs font-medium text-foam/70">{{"{{"}} t "home.next_steps" {{"}}"}}</p>
        <span class="font-mono text-[11px] text-foam/50">{{"{{"}} t "home.stack" {{"}}"}}</span>
      </div>
      <ol class="space-y-3 px-5 py-5 text-sm text-foam/70">
        <li class="flex gap-3"><span class="font-mono text-xs leading-5 tabular-nums text-foam/50">01</span><span>{{"{{"}} t "home.step_resource" {{"}}"}}</span></li>
        <li class="flex gap-3"><span class="font-mono text-xs leading-5 tabular-nums text-foam/50">02</span><span>{{"{{"}} t "home.step_dev" {{"}}"}}</span></li>
        <li class="flex gap-3"><span class="font-mono text-xs leading-5 tabular-nums text-foam/50">03</span><span>{{"{{"}} t "home.step_docs" {{"}}"}}</span></li>
      </ol>
      <div class="border-t border-foam/10 bg-ink/60 px-5 py-4">
        <pre class="font-mono text-[13px] leading-6 text-copper"><span class="select-none text-foam/40">$ </span>amarra-cais g resource bookmark
<span class="select-none text-foam/40">  </span>--fields title:string --public</pre>
      </div>
    </div>
    <p class="mt-3 px-1 text-xs text-foam/50">{{"{{"}} t "home.powered_by" {{"}}"}}</p>
  </aside>
</section>
{{"{{"}} end {{"}}"}}
`

const tplPageHome = tplPageHomePre + tplPageHomeCTA + tplPageHomePost

const tplPageHomeSlim = tplPageHomePre + tplPageHomePost

const tplPageDashboard = `{{"{{"}} define "content" {{"}}"}}
<div class="mx-auto w-full max-w-4xl">
  <h1 class="font-serif text-3xl font-medium text-foam">{{"{{"}} t "dashboard.title" {{"}}"}}</h1>
  <div class="mt-6 grid gap-4 sm:grid-cols-2">
    <.stat label="{{"{{"}} t "dashboard.env" {{"}}"}}" value="{{"{{"}} .Env {{"}}"}}" />
    <.stat label="{{"{{"}} t "dashboard.locale" {{"}}"}}" value="{{"{{"}} .Locale {{"}}"}}" />
  </div>
  <.form action="/logout" method="post" class="mt-8">
    <.button type="submit" class="amarra-btn amarra-btn-secondary">{{"{{"}} t "auth.logout" {{"}}"}}</.button>
  </.form>
</div>
{{"{{"}} end {{"}}"}}
`
