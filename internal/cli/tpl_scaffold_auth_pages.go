// Auth page templates for cais g auth and cais new (full app).
package cli

const tplPageLogin = `{{"{{"}} define "content" {{"}}"}}
<div class="flex min-h-[calc(100svh-11rem)] items-center justify-center py-8">
  <div class="w-full max-w-sm rounded-xl border border-foam/10 bg-tide/40 p-6">
    <h1 class="font-serif text-2xl font-medium text-foam">{{"{{"}} t "auth.login_title" {{"}}"}}</h1>
    <.form action="/login" method="post" class="mt-6 space-y-5">
      <.input name="email" type="email" label="Email" value="{{"{{"}} .Email {{"}}"}}" error="{{"{{"}} fieldError .Errors "email" {{"}}"}}" />
      <.password name="password" label="{{"{{"}} t "auth.password_label" {{"}}"}}" error="{{"{{"}} fieldError .Errors "password" {{"}}"}}" autocomplete="current-password" required="required" />
      <.button type="submit" class="amarra-btn amarra-btn-primary w-full">{{"{{"}} t "auth.login_submit" {{"}}"}}</.button>
    </.form>
    <div class="mt-6 space-y-2 text-center text-sm text-foam/60">
      <p>
        {{"{{"}} t "auth.signup_prompt" {{"}}"}}
        <a class="text-foam underline decoration-foam/40 underline-offset-4 hover:decoration-foam/70" href="/signup">{{"{{"}} t "auth.signup_title" {{"}}"}}</a>
      </p>
      <p><a class="text-foam underline decoration-foam/40 underline-offset-4 hover:decoration-foam/70" href="/forgot-password">{{"{{"}} t "auth.forgot_password" {{"}}"}}</a></p>
    </div>
  </div>
</div>
{{"{{"}} end {{"}}"}}
`

const tplPageSignup = `{{"{{"}} define "content" {{"}}"}}
<div class="flex min-h-[calc(100svh-11rem)] items-center justify-center py-8">
  <div class="w-full max-w-sm rounded-xl border border-foam/10 bg-tide/40 p-6">
    <h1 class="font-serif text-2xl font-medium text-foam">{{"{{"}} t "auth.signup_title" {{"}}"}}</h1>
    <.form action="/signup" method="post" class="mt-6 space-y-5">
      <.input name="email" type="email" label="{{"{{"}} t "auth.email_label" {{"}}"}}" value="{{"{{"}} .Email {{"}}"}}" error="{{"{{"}} fieldError .Errors "email" {{"}}"}}" />
      <.password name="password" label="{{"{{"}} t "auth.password_label" {{"}}"}}" error="{{"{{"}} fieldError .Errors "password" {{"}}"}}" autocomplete="new-password" required="required" />
      <.password name="password_confirmation" label="{{"{{"}} t "auth.password_confirmation_label" {{"}}"}}" error="{{"{{"}} fieldError .Errors "password_confirmation" {{"}}"}}" autocomplete="new-password" required="required" />
      <.button type="submit" class="amarra-btn amarra-btn-primary w-full">{{"{{"}} t "auth.signup_submit" {{"}}"}}</.button>
    </.form>
    <p class="mt-6 text-center text-sm text-foam/60">
      {{"{{"}} t "auth.login_prompt" {{"}}"}}
      <a class="text-foam underline decoration-foam/40 underline-offset-4 hover:decoration-foam/70" href="/login">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
    </p>
  </div>
</div>
{{"{{"}} end {{"}}"}}`

const tplPageForgotPassword = `{{"{{"}} define "content" {{"}}"}}
<div class="flex min-h-[calc(100svh-11rem)] items-center justify-center py-8">
  <div class="w-full max-w-sm rounded-xl border border-foam/10 bg-tide/40 p-6">
    <h1 class="font-serif text-2xl font-medium text-foam">{{"{{"}} t "auth.forgot_password_title" {{"}}"}}</h1>
    <p class="mt-2 text-sm text-foam/60">{{"{{"}} t "auth.forgot_password_help" {{"}}"}}</p>
    <.form action="/forgot-password" method="post" class="mt-6 space-y-5">
      <.input name="email" type="email" label="{{"{{"}} t "auth.email_label" {{"}}"}}" value="{{"{{"}} .Email {{"}}"}}" error="{{"{{"}} fieldError .Errors "email" {{"}}"}}" />
      <.button type="submit" class="amarra-btn amarra-btn-primary w-full">{{"{{"}} t "auth.forgot_password_submit" {{"}}"}}</.button>
    </.form>
    <p class="mt-6 text-center text-sm text-foam/60">
      <a class="text-foam underline decoration-foam/40 underline-offset-4 hover:decoration-foam/70" href="/login">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
    </p>
  </div>
</div>
{{"{{"}} end {{"}}"}}`

const tplPageResetPassword = `{{"{{"}} define "content" {{"}}"}}
<div class="flex min-h-[calc(100svh-11rem)] items-center justify-center py-8">
  <div class="w-full max-w-sm rounded-xl border border-foam/10 bg-tide/40 p-6">
    <h1 class="font-serif text-2xl font-medium text-foam">{{"{{"}} t "auth.reset_password_title" {{"}}"}}</h1>
    {{"{{"}} if .Errors {{"}}"}}<div class="mt-3 text-sm text-danger">{{"{{"}} fieldError .Errors "token" {{"}}"}}</div>{{"{{"}} end {{"}}"}}
    <.form action="/reset-password" method="post" class="mt-6 space-y-5">
      <input type="hidden" name="token" value="{{"{{"}} .Token {{"}}"}}" />
      <.password name="password" label="{{"{{"}} t "auth.password_label" {{"}}"}}" error="{{"{{"}} fieldError .Errors "password" {{"}}"}}" autocomplete="new-password" required="required" />
      <.password name="password_confirmation" label="{{"{{"}} t "auth.password_confirmation_label" {{"}}"}}" error="{{"{{"}} fieldError .Errors "password_confirmation" {{"}}"}}" autocomplete="new-password" required="required" />
      <.button type="submit" class="amarra-btn amarra-btn-primary w-full">{{"{{"}} t "auth.reset_password_submit" {{"}}"}}</.button>
    </.form>
    <p class="mt-6 text-center text-sm text-foam/60">
      <a class="text-foam underline decoration-foam/40 underline-offset-4 hover:decoration-foam/70" href="/login">{{"{{"}} t "auth.login_title" {{"}}"}}</a>
    </p>
  </div>
</div>
{{"{{"}} end {{"}}"}}`

const tplPageSettings = `{{"{{"}} define "content" {{"}}"}}
<div class="flex min-h-[calc(100svh-11rem)] items-center justify-center py-8">
  <div class="w-full max-w-sm rounded-xl border border-foam/10 bg-tide/40 p-6">
    <h1 class="font-serif text-2xl font-medium text-foam">{{"{{"}} t "settings.title" {{"}}"}}</h1>
    <.form action="/settings" method="post" data-amarra-skip class="mt-6 space-y-5">
      <.password name="current_password" label="{{"{{"}} t "settings.current_password" {{"}}"}}" error="{{"{{"}} fieldError .Errors "current_password" {{"}}"}}" autocomplete="current-password" required="required" />
      <.password name="password" label="{{"{{"}} t "auth.password_label" {{"}}"}}" error="{{"{{"}} fieldError .Errors "password" {{"}}"}}" autocomplete="new-password" required="required" />
      <.password name="password_confirmation" label="{{"{{"}} t "auth.password_confirmation_label" {{"}}"}}" error="{{"{{"}} fieldError .Errors "password_confirmation" {{"}}"}}" autocomplete="new-password" required="required" />
      <.button type="submit" class="amarra-btn amarra-btn-primary w-full">{{"{{"}} t "settings.submit" {{"}}"}}</.button>
    </.form>
  </div>
</div>
{{"{{"}} end {{"}}"}}`
