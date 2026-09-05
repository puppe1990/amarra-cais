// Auth page templates for cais g auth and cais new (full app).
package cli

const tplPageLogin = `{{"{{"}} define "content" {{"}}"}}
<div class="flex min-h-screen items-center justify-center px-4 py-10">
  <div class="w-full max-w-md mx-auto p-4 md:p-6 bg-white rounded-2xl border border-slate-200/80 shadow-lg">
    <div class="mb-4 pb-4 border-b border-slate-100">
      <h2 class="text-lg font-black tracking-tight text-slate-900 font-display">{{"{{"}} t "auth.login_title" {{"}}"}}</h2>
    </div>
    <.form action="/login" method="post">
      {{"{{"}} csrfField .CSRFToken {{"}}"}}
      <.input name="email" type="email" label="Email" value="{{"{{"}} .Email {{"}}"}}" error="{{"{{"}} fieldError .Errors "email" {{"}}"}}" />
      {{"{{"}} fieldPassword (makeField "password" (t "auth.password_label") "" "password" true .Errors) {{"}}"}}
      <.button type="submit">{{"{{"}} t "auth.login_submit" {{"}}"}}</.button>
    </.form>
    <p class="text-sm text-slate-600 mt-4 text-center space-y-1">
      <span class="block">
        {{"{{"}} t "auth.signup_prompt" {{"}}"}}
        <a class="text-indigo-600 hover:text-indigo-800" href="/signup" data-amarra-drive="true">{{"{{"}} t "auth.signup_title" {{"}}"}}</a>
      </span>
      <a class="text-indigo-600 hover:text-indigo-800" href="/forgot-password" data-amarra-drive="true">{{"{{"}} t "auth.forgot_password" {{"}}"}}</a>
    </p>
  </div>
</div>
{{"{{"}} end {{"}}"}}
`

const tplPageSignup = `{{"{{"}} define "content" {{"}}"}}
<div class="flex min-h-screen items-center justify-center px-4 py-10">
  <div class="w-full max-w-md mx-auto p-4 md:p-6 bg-white rounded-2xl border border-slate-200/80 shadow-lg">
    <div class="mb-4 pb-4 border-b border-slate-100">
      <h2 class="text-lg font-black tracking-tight text-slate-900 font-display">{{"{{"}} t "auth.signup_title" {{"}}"}}</h2>
    </div>
    <.form action="/signup" method="post">
      {{"{{"}} csrfField .CSRFToken {{"}}"}}
      <.input name="email" type="email" label="{{"{{"}} t "contact.email_label" {{"}}"}}" value="{{"{{"}} .Email {{"}}"}}" error="{{"{{"}} fieldError .Errors "email" {{"}}"}}" />
      {{"{{"}} fieldPassword (makeField "password" (t "auth.password_label") "" "password" true .Errors) {{"}}"}}
      {{"{{"}} fieldPassword (makeField "password_confirmation" (t "auth.password_confirmation_label") "" "password" true .Errors) {{"}}"}}
      <.button type="submit">{{"{{"}} t "auth.signup_submit" {{"}}"}}</.button>
    </.form>
    <p class="text-sm text-slate-600 mt-4 text-center">
      {{"{{"}} t "auth.login_prompt" {{"}}"}}
      <a class="text-indigo-600 hover:text-indigo-800" href="/login" data-amarra-drive="true">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
    </p>
  </div>
</div>
{{"{{"}} end {{"}}"}}`

const tplPageForgotPassword = `{{"{{"}} define "content" {{"}}"}}
<div class="flex min-h-screen items-center justify-center px-4 py-10">
  <div class="w-full max-w-md mx-auto p-4 md:p-6 bg-white rounded-2xl border border-slate-200/80 shadow-lg">
    <div class="mb-4 pb-4 border-b border-slate-100">
      <h2 class="text-lg font-black tracking-tight text-slate-900 font-display">{{"{{"}} t "auth.forgot_password_title" {{"}}"}}</h2>
      <p class="text-[11px] text-slate-500 font-medium mt-1">{{"{{"}} t "auth.forgot_password_help" {{"}}"}}</p>
    </div>
    <.form action="/forgot-password" method="post">
      {{"{{"}} csrfField .CSRFToken {{"}}"}}
      <.input name="email" type="email" label="{{"{{"}} t "contact.email_label" {{"}}"}}" value="{{"{{"}} .Email {{"}}"}}" error="{{"{{"}} fieldError .Errors "email" {{"}}"}}" />
      <.button type="submit">{{"{{"}} t "auth.forgot_password_submit" {{"}}"}}</.button>
    </.form>
    <p class="text-sm text-slate-600 mt-4 text-center">
      <a class="text-indigo-600 hover:text-indigo-800" href="/login" data-amarra-drive="true">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
    </p>
  </div>
</div>
{{"{{"}} end {{"}}"}}`

const tplPageResetPassword = `{{"{{"}} define "content" {{"}}"}}
<div class="flex min-h-screen items-center justify-center px-4 py-10">
  <div class="w-full max-w-md mx-auto p-4 md:p-6 bg-white rounded-2xl border border-slate-200/80 shadow-lg">
    <div class="mb-4 pb-4 border-b border-slate-100">
      <h2 class="text-lg font-black tracking-tight text-slate-900 font-display">{{"{{"}} t "auth.reset_password_title" {{"}}"}}</h2>
    </div>
    {{"{{"}} if .Errors {{"}}"}}{{"{{"}} fieldError .Errors "token" {{"}}"}}{{"{{"}} end {{"}}"}}
    <.form action="/reset-password" method="post">
      {{"{{"}} csrfField .CSRFToken {{"}}"}}
      <input type="hidden" name="token" value="{{"{{"}} .Token {{"}}"}}" />
      {{"{{"}} fieldPassword (makeField "password" (t "auth.password_label") "" "password" true .Errors) {{"}}"}}
      {{"{{"}} fieldPassword (makeField "password_confirmation" (t "auth.password_confirmation_label") "" "password" true .Errors) {{"}}"}}
      <.button type="submit">{{"{{"}} t "auth.reset_password_submit" {{"}}"}}</.button>
    </.form>
    <p class="text-sm text-slate-600 mt-4 text-center">
      <a class="text-indigo-600 hover:text-indigo-800" href="/login" data-amarra-drive="true">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
    </p>
  </div>
</div>
{{"{{"}} end {{"}}"}}`
