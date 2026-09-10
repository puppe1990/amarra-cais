// Auth page templates for cais g auth and cais new (full app).
package cli

const tplPageLogin = `{{"{{"}} define "content" {{"}}"}}
<div class="flex items-center justify-center px-4 py-16">
<div class="w-full max-w-md mx-auto border border-copper/45 bg-ink/70 p-6 md:p-7 shadow-[10px_10px_0_0_rgba(201,137,58,0.22)]">
  <h2 class="font-serif text-2xl text-foam">{{"{{"}} t "auth.login_title" {{"}}"}}</h2>
  <.form action="/login" method="post">
    {{"{{"}} csrfField .CSRFToken {{"}}"}}
    <.input name="email" type="email" label="Email" value="{{"{{"}} .Email {{"}}"}}" error="{{"{{"}} fieldError .Errors "email" {{"}}"}}" />
    <.password name="password" label="{{"{{"}} t "auth.password_label" {{"}}"}}" error="{{"{{"}} fieldError .Errors "password" {{"}}"}}" autocomplete="current-password" required="required" />
    <.button type="submit">{{"{{"}} t "auth.login_submit" {{"}}"}}</.button>
  </.form>
  <p class="mt-4 text-center text-sm text-foam/60 space-y-1">
    <span class="block">
      {{"{{"}} t "auth.signup_prompt" {{"}}"}}
      <a class="text-copper hover:text-foam" href="/signup">{{"{{"}} t "auth.signup_title" {{"}}"}}</a>
    </span>
    <a class="text-copper hover:text-foam" href="/forgot-password">{{"{{"}} t "auth.forgot_password" {{"}}"}}</a>
  </p>
</div>
</div>
{{"{{"}} end {{"}}"}}
`

const tplPageSignup = `{{"{{"}} define "content" {{"}}"}}
<div class="flex items-center justify-center px-4 py-16">
<div class="w-full max-w-md mx-auto border border-copper/45 bg-ink/70 p-6 md:p-7 shadow-[10px_10px_0_0_rgba(201,137,58,0.22)]">
  <h2 class="font-serif text-2xl text-foam">{{"{{"}} t "auth.signup_title" {{"}}"}}</h2>
  <.form action="/signup" method="post">
    {{"{{"}} csrfField .CSRFToken {{"}}"}}
    <.input name="email" type="email" label="{{"{{"}} t "contact.email_label" {{"}}"}}" value="{{"{{"}} .Email {{"}}"}}" error="{{"{{"}} fieldError .Errors "email" {{"}}"}}" />
    <.password name="password" label="{{"{{"}} t "auth.password_label" {{"}}"}}" error="{{"{{"}} fieldError .Errors "password" {{"}}"}}" autocomplete="new-password" required="required" />
    <.password name="password_confirmation" label="{{"{{"}} t "auth.password_confirmation_label" {{"}}"}}" error="{{"{{"}} fieldError .Errors "password_confirmation" {{"}}"}}" autocomplete="new-password" required="required" />
    <.button type="submit">{{"{{"}} t "auth.signup_submit" {{"}}"}}</.button>
  </.form>
  <p class="mt-4 text-center text-sm text-foam/60">
    {{"{{"}} t "auth.login_prompt" {{"}}"}}
    <a class="text-copper hover:text-foam" href="/login">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
  </p>
</div>
</div>
{{"{{"}} end {{"}}"}}`

const tplPageForgotPassword = `{{"{{"}} define "content" {{"}}"}}
<div class="flex items-center justify-center px-4 py-16">
<div class="w-full max-w-md mx-auto border border-copper/45 bg-ink/70 p-6 md:p-7 shadow-[10px_10px_0_0_rgba(201,137,58,0.22)]">
  <h2 class="font-serif text-2xl text-foam">{{"{{"}} t "auth.forgot_password_title" {{"}}"}}</h2>
  <p class="mt-1 font-mono text-[11px] uppercase tracking-[0.22em] text-copper">{{"{{"}} t "auth.forgot_password_help" {{"}}"}}</p>
  <.form action="/forgot-password" method="post">
    {{"{{"}} csrfField .CSRFToken {{"}}"}}
    <.input name="email" type="email" label="{{"{{"}} t "contact.email_label" {{"}}"}}" value="{{"{{"}} .Email {{"}}"}}" error="{{"{{"}} fieldError .Errors "email" {{"}}"}}" />
    <.button type="submit">{{"{{"}} t "auth.forgot_password_submit" {{"}}"}}</.button>
  </.form>
  <p class="mt-4 text-center text-sm text-foam/60">
    <a class="text-copper hover:text-foam" href="/login">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
  </p>
</div>
</div>
{{"{{"}} end {{"}}"}}`

const tplPageResetPassword = `{{"{{"}} define "content" {{"}}"}}
<div class="flex items-center justify-center px-4 py-16">
<div class="w-full max-w-md mx-auto border border-copper/45 bg-ink/70 p-6 md:p-7 shadow-[10px_10px_0_0_rgba(201,137,58,0.22)]">
  <h2 class="font-serif text-2xl text-foam">{{"{{"}} t "auth.reset_password_title" {{"}}"}}</h2>
  {{"{{"}} if .Errors {{"}}"}}{{"{{"}} fieldError .Errors "token" {{"}}"}}{{"{{"}} end {{"}}"}}
  <.form action="/reset-password" method="post">
    {{"{{"}} csrfField .CSRFToken {{"}}"}}
    <input type="hidden" name="token" value="{{"{{"}} .Token {{"}}"}}" />
    <.password name="password" label="{{"{{"}} t "auth.password_label" {{"}}"}}" error="{{"{{"}} fieldError .Errors "password" {{"}}"}}" autocomplete="new-password" required="required" />
    <.password name="password_confirmation" label="{{"{{"}} t "auth.password_confirmation_label" {{"}}"}}" error="{{"{{"}} fieldError .Errors "password_confirmation" {{"}}"}}" autocomplete="new-password" required="required" />
    <.button type="submit">{{"{{"}} t "auth.reset_password_submit" {{"}}"}}</.button>
  </.form>
  <p class="mt-4 text-center text-sm text-foam/60">
    <a class="text-copper hover:text-foam" href="/login">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
  </p>
</div>
</div>
{{"{{"}} end {{"}}"}}`
