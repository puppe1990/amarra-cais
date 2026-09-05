package cli

// Amarra layout for cais new (full, minimal, blank). One shell: #amarra-main + amarra.js.
const tplLayoutTitleDesc = `{{"{{"}} define "title" {{"}}"}}{{.AppName}}{{"{{"}} end {{"}}"}}
{{"{{"}} define "description" {{"}}"}}{{.AppName}} — powered by Amarra{{"{{"}} end {{"}}"}}`

const tplLayoutBaseOpen = `{{"{{"}} define "app" {{"}}"}}
<!doctype html>
<html lang="{{"{{"}} htmlLang {{"}}"}}">
  <head>
    <meta charset="UTF-8" />
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
    <meta name="theme-color" content="#4f46e5" />
    <meta name="mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-capable" content="yes" />
    <meta name="apple-mobile-web-app-status-bar-style" content="black-translucent" />
    <meta name="apple-mobile-web-app-title" content="{{.AppName}}" />
    <link rel="apple-touch-icon" href="/static/icons/icon.png" />
    <link rel="icon" href="/static/icons/icon.png" type="image/png" />
    <script src="/static/js/amarra.js" defer></script>
  </head>
  <body class="min-h-screen bg-slate-50 font-sans antialiased text-slate-900 flex flex-col justify-between">
    <div>
      <header class="bg-white border-b border-slate-200 sticky top-0 z-40 shadow-xs">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-2.5 flex flex-col md:flex-row md:items-center md:justify-between gap-3">
          <a href="/" class="flex items-center gap-2.5 group" data-amarra-drive="true">
            <div class="p-2 bg-indigo-600 rounded-lg text-white shadow-xs flex items-center justify-center group-hover:bg-indigo-700 transition">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
              </svg>
            </div>
            <div>
              <h1 class="text-lg font-black text-slate-900 tracking-tight font-display flex items-center gap-1.5 leading-none">
                {{.AppName}}
                <span class="text-[9px] bg-indigo-100 text-indigo-800 px-1.5 py-0.5 rounded-md font-bold uppercase tracking-wider">Beta</span>
              </h1>
              <p class="text-[10px] text-slate-500 font-semibold mt-1">Powered by Amarra</p>
            </div>
          </a>
        </div>
      </header>
      <nav id="amarra-nav" class="bg-white border-b border-slate-200 shadow-2xs sticky top-[53px] z-30">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div class="flex space-x-1 py-1.5 overflow-x-auto no-scrollbar">
            `

const tplLayoutNavFull = `<!-- cais:nav -->
            {{"{{"}} template "nav_links" . {{"}}"}}`

const tplLayoutNavEmpty = `<!-- cais:nav -->`

const tplLayoutBaseClose = `
          </div>
        </div>
      </nav>
      <div id="amarra-toast-host" aria-live="polite">
        <.flash />
      </div>
      <main id="amarra-main" class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-5 flex-grow">{{"{{"}} template "content" . {{"}}"}}</main>
    </div>
    <footer class="mt-auto border-t border-slate-200/80 pt-8 pb-6 text-center text-xs text-slate-400">
      <div class="max-w-7xl mx-auto px-4">
        <p>© 2026 {{.AppName}}. Built with Amarra.</p>
        <p class="mt-1">HTML + Go + SQLite — server-rendered, app-like UX.</p>
      </div>
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
<div class="flex flex-col items-center justify-center px-6 py-14 text-center">
  <h1 class="mt-10 font-serif text-4xl font-semibold tracking-tight text-stone-800 md:text-5xl">{{"{{"}} t "home.rails_heading" {{"}}"}}</h1>
  <p data-testid="amarra-ready" class="mt-3 text-lg text-stone-600">{{"{{"}} t "home.rails_subtitle" .Site.AppName {{"}}"}}</p>
  <p class="mt-2 text-sm text-stone-500">{{"{{"}} t "home.stack" {{"}}"}}</p>
</div>
{{"{{"}} end {{"}}"}}
`

const tplPageContact = `{{"{{"}} define "content" {{"}}"}}
<div class="w-full max-w-md mx-auto p-4 md:p-5 bg-white rounded-xl border border-slate-200 shadow-2xs">
  <div class="mb-4 pb-4 border-b border-slate-100">
    <h2 class="text-lg font-black tracking-tight text-slate-900 font-display">{{"{{"}} t "contact.heading" {{"}}"}}</h2>
    <p class="text-[11px] text-slate-500 font-medium mt-1">{{"{{"}} t "contact.title" {{"}}"}}</p>
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
<div class="w-full max-w-4xl mx-auto p-4 md:p-5 bg-white rounded-xl border border-slate-200 shadow-2xs">
  <div class="mb-4 pb-4 border-b border-slate-100">
    <h2 class="text-lg font-black tracking-tight text-slate-900 font-display">{{"{{"}} t "dashboard.title" {{"}}"}}</h2>
  </div>
  <p>{{"{{"}} t "dashboard.contacts" {{"}}"}} {{"{{"}} .TotalContacts {{"}}"}}</p>
  <p>{{"{{"}} t "dashboard.env" {{"}}"}} {{"{{"}} .Env {{"}}"}}</p>
  <.form action="/logout" method="post">
    {{"{{"}} csrfField .CSRFToken {{"}}"}}
    <.button type="submit">{{"{{"}} t "auth.logout" {{"}}"}}</.button>
  </.form>
</div>
{{"{{"}} end {{"}}"}}
`
