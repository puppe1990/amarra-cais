package cli

// Amarra layout for cais new (full, minimal, blank). One shell: #amarra-main + amarra.js.
const tplLayoutTitleDesc = `{{"{{"}} define "title" {{"}}"}}{{.AppName}}{{"{{"}} end {{"}}"}}
{{"{{"}} define "description" {{"}}"}}{{.AppName}} — powered by Amarra{{"{{"}} end {{"}}"}}`

const tplLayoutBaseOpen = `{{"{{"}} define "app" {{"}}"}}
<!doctype html>
<html lang="{{"{{"}} if .HTMLLang {{"}}"}}{{"{{"}} .HTMLLang {{"}}"}}{{"{{"}} else {{"}}"}}{{"{{"}} htmlLang {{"}}"}}{{"{{"}} end {{"}}"}}" data-amarra-layout="app">
  <head>
    <meta charset="UTF-8" />
    <script>
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
    <link rel="stylesheet" href="/static/css/styles.css" />
    <link rel="manifest" href="/static/manifest.webmanifest" />
    <meta name="theme-color" content="#c9893a" />
    <meta name="mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
    <meta name="apple-mobile-web-app-title" content="{{.AppName}}" />
    <link rel="apple-touch-icon" href="/static/icons/icon.png" />
    <link rel="icon" href="/static/icons/icon.png" type="image/png" />
    <script src="/static/js/amarra.js" defer></script>
  </head>
  <body class="min-h-screen bg-ink font-sans antialiased text-foam flex flex-col justify-between">
    <div>
      <header class="bg-ink/95 backdrop-blur-sm border-b border-copper/30 sticky top-0 z-40">
        <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-3 flex items-center justify-between gap-4">
          <a href="/" class="flex items-center gap-3 group">
            <span class="flex h-9 w-9 items-center justify-center border border-copper/70 text-copper group-hover:bg-copper group-hover:text-ink transition-colors" aria-hidden="true">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.75" d="M5 20V8m0 0a3 3 0 116 0v1H5m14 11V9a3 3 0 00-3-3h-3" /></svg>
            </span>
            <span>
              <span class="block font-serif text-xl leading-none tracking-tight text-foam">{{.AppName}}</span>
              <span class="mt-1 block font-mono text-[10px] uppercase tracking-[0.28em] text-copper/80">Amarra</span>
            </span>
          </a>
          <a href="/login" class="font-mono text-[10px] uppercase tracking-[0.22em] text-copper hover:text-foam transition-colors">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
        </div>
      </header>
      <nav id="amarra-nav" class="bg-ink border-b border-foam/10 sticky top-[57px] z-30">
        <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
          <!-- nav hook re-syncs active link after Drive morph (#27); SSR ActiveNav stays the first-paint default -->
          <div amarra-hook="nav" data-amarra-nav-on="text-copper" data-amarra-nav-off="text-foam/50 hover:text-foam" class="flex gap-1 py-1 overflow-x-auto no-scrollbar">
            `

const tplLayoutNavFull = `<!-- cais:nav -->
            {{"{{"}} template "nav_links" . {{"}}"}}
            <.locale-toggle current="{{"{{"}} .Locale {{"}}"}}" />`

const tplLayoutNavEmpty = `<!-- cais:nav -->`

const tplLayoutBaseClose = `
          </div>
        </div>
      </nav>
      <div id="amarra-toast-host" aria-live="polite"></div>
      <main id="amarra-main" class="flex-grow px-4 sm:px-6 lg:px-8 py-5">
        <.flash />
        {{"{{"}} template "content" . {{"}}"}}
      </main>
    </div>
    <footer class="mt-auto border-t border-foam/10 py-6 text-center font-mono text-[10px] uppercase tracking-[0.22em] text-foam/40">
      <p>© 2026 {{.AppName}} · Built with Amarra</p>
    </footer>
    {{"{{"}} if eq .Site.Env "development" {{"}}"}}
    <script>
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
    </script>
    {{"{{"}} else {{"}}"}}
    <script>
      if ("serviceWorker" in navigator) {
        navigator.serviceWorker.register("/static/js/sw.js");
      }
    </script>
    {{"{{"}} end {{"}}"}}
  </body>
</html>
{{"{{"}} end {{"}}"}}`

const tplLayout = tplLayoutTitleDesc + tplPartialIcons + tplPartialNavLinks + tplLayoutBaseOpen + tplLayoutNavFull + tplLayoutBaseClose

const tplLayoutMinimal = tplLayoutTitleDesc + tplPartialIcons + tplPartialNavLinks + tplLayoutBaseOpen + tplLayoutNavEmpty + tplLayoutBaseClose

const tplLayoutBlank = tplLayoutMinimal

const tplLayoutWelcome = tplLayoutMinimal

const tplCaisLogo = `{{"{{"}} define "cais_logo" {{"}}"}}
<img
  src="/static/img/go-on-cais.jpg"
  alt="Go on Cais"
  width="1024"
  height="683"
  class="w-full max-w-lg rounded-2xl shadow-xl shadow-amber-950/15 ring-1 ring-amber-900/10"
/>
{{"{{"}} end {{"}}"}}
`

const tplPageHome = `{{"{{"}} define "content" {{"}}"}}
<section class="amarra-grain relative isolate overflow-hidden min-h-[calc(100vh-12rem)] -mx-4 sm:-mx-6 lg:-mx-8 -my-5 px-6 sm:px-10 lg:px-14 py-16 md:py-24">
  <div class="pointer-events-none absolute -top-24 left-[12%] h-80 w-80 rounded-full bg-copper/30 blur-3xl amarra-lantern"></div>
  <div class="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_12%_-10%,rgba(201,137,58,0.22),transparent_42%),radial-gradient(ellipse_at_100%_110%,rgba(18,70,78,0.55),transparent_50%)]"></div>
  <svg class="pointer-events-none absolute right-[-6%] top-4 hidden h-[28rem] w-[28rem] text-copper/15 lg:block" viewBox="0 0 200 200" fill="none" aria-hidden="true">
    <circle cx="100" cy="100" r="78" stroke="currentColor" stroke-width="0.6"></circle>
    <circle cx="100" cy="100" r="52" stroke="currentColor" stroke-width="0.6"></circle>
    <path d="M100 22v156M22 100h156" stroke="currentColor" stroke-width="0.4"></path>
    <rect x="86" y="68" width="28" height="72" rx="4" stroke="currentColor" stroke-width="1.5"></rect>
    <path d="M76 92h48M76 108h48" stroke="currentColor" stroke-width="2.2"></path>
  </svg>
  <div class="relative mx-auto max-w-6xl grid items-end gap-14 lg:grid-cols-12">
    <div class="lg:col-span-7">
      <p class="amarra-rise font-mono text-[11px] uppercase tracking-[0.32em] text-copper">{{"{{"}} t "home.berth_label" {{"}}"}}</p>
      <h1 class="amarra-rise amarra-rise-delay-1 mt-5 font-serif italic text-[clamp(2.75rem,7vw,6.25rem)] leading-[0.88] text-foam">{{"{{"}} t "home.rails_heading" {{"}}"}}</h1>
      <p data-testid="amarra-ready" class="amarra-rise amarra-rise-delay-2 mt-8 max-w-md text-lg leading-relaxed text-foam/65">{{"{{"}} t "home.rails_subtitle" .Site.AppName {{"}}"}}</p>
      <div class="amarra-rise amarra-rise-delay-3 mt-10 flex flex-wrap items-center gap-6">
        <a href="/login" class="inline-flex items-center gap-2 bg-copper px-5 py-2.5 font-mono text-[11px] uppercase tracking-[0.22em] text-ink hover:bg-foam transition-colors">{{"{{"}} t "home.cta_board" {{"}}"}}</a>
        <a href="/contact" class="font-mono text-[11px] uppercase tracking-[0.22em] text-foam/55 hover:text-copper transition-colors">{{"{{"}} t "home.contact_link" {{"}}"}}</a>
      </div>
      <div class="amarra-rise amarra-rise-delay-4 mt-14 flex items-center gap-4 text-copper/50">
        <span class="h-px w-16 bg-copper/60"></span>
        <span class="font-mono text-[10px] uppercase tracking-[0.35em]">{{"{{"}} t "home.tide_mark" {{"}}"}}</span>
      </div>
    </div>
    <aside class="lg:col-span-5 amarra-rise amarra-rise-delay-2">
      <div class="border border-copper/45 bg-ink/70 p-7 shadow-[10px_10px_0_0_rgba(201,137,58,0.22)]">
        <p class="font-mono text-[10px] uppercase tracking-[0.26em] text-copper">{{"{{"}} t "home.manifest_label" {{"}}"}}</p>
        <p class="mt-3 font-serif text-3xl text-foam">{{"{{"}} .Site.AppName {{"}}"}}</p>
        <p class="mt-1 font-mono text-xs text-foam/45">{{"{{"}} t "home.stack" {{"}}"}}</p>
        <ol class="mt-7 space-y-3 font-mono text-[12px] text-foam/75">
          <li><span class="text-copper">01 —</span> {{"{{"}} t "home.step_resource" {{"}}"}}</li>
          <li><span class="text-copper">02 —</span> {{"{{"}} t "home.step_dev" {{"}}"}}</li>
          <li><span class="text-copper">03 —</span> {{"{{"}} t "home.step_docs" {{"}}"}}</li>
        </ol>
        <pre class="mt-7 whitespace-pre-wrap break-words border-t border-foam/10 pt-4 font-mono text-[11px] leading-relaxed text-copper">amarra-cais g resource bookmark
--fields title:string --public</pre>
      </div>
    </aside>
  </div>
</section>
{{"{{"}} end {{"}}"}}
`

const tplPageContact = `{{"{{"}} define "content" {{"}}"}}
<div class="w-full max-w-md mx-auto border border-copper/45 bg-ink/70 p-6 md:p-7 shadow-[10px_10px_0_0_rgba(201,137,58,0.22)]">
  <div class="mb-5 pb-4 border-b border-foam/10">
    <h2 class="font-serif text-2xl text-foam">{{"{{"}} t "contact.heading" {{"}}"}}</h2>
    <p class="mt-1 font-mono text-[11px] uppercase tracking-[0.22em] text-copper">{{"{{"}} t "contact.title" {{"}}"}}</p>
  </div>
  <.form action="/contact" method="post">
    {{"{{"}} csrfField .CSRFToken {{"}}"}}
    <.input name="name" label="{{"{{"}} t "contact.name_label" {{"}}"}}" value="{{"{{"}} .Name {{"}}"}}" error="{{"{{"}} fieldError .Errors "name" {{"}}"}}" />
    <.input name="email" type="email" label="{{"{{"}} t "contact.email_label" {{"}}"}}" value="{{"{{"}} .Email {{"}}"}}" error="{{"{{"}} fieldError .Errors "email" {{"}}"}}" />
    <.button type="submit">{{"{{"}} t "contact.submit" {{"}}"}}</.button>
  </.form>
</div>
{{"{{"}} end {{"}}"}}
`

const tplPageDashboard = `{{"{{"}} define "content" {{"}}"}}
<div class="w-full max-w-4xl mx-auto border border-copper/45 bg-ink/70 p-6 md:p-7 shadow-[10px_10px_0_0_rgba(201,137,58,0.22)]">
  <div class="mb-5 pb-4 border-b border-foam/10">
    <h2 class="font-serif text-2xl text-foam">{{"{{"}} t "dashboard.title" {{"}}"}}</h2>
  </div>
  <p class="text-foam/80">{{"{{"}} t "dashboard.contacts" {{"}}"}} {{"{{"}} .TotalContacts {{"}}"}}</p>
  <p class="mt-2 text-foam/80">{{"{{"}} t "dashboard.env" {{"}}"}} {{"{{"}} .Env {{"}}"}}</p>
  <.form action="/logout" method="post">
    {{"{{"}} csrfField .CSRFToken {{"}}"}}
    <.button type="submit">{{"{{"}} t "auth.logout" {{"}}"}}</.button>
  </.form>
</div>
{{"{{"}} end {{"}}"}}
`
