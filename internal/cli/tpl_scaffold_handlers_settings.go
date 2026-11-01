package cli

const tplAuthSettings = `package handlers

import (
	"net/http"

	"github.com/puppe1990/amarra-cais/pkg/cais/flash"
	"github.com/puppe1990/amarra-cais/pkg/cais/httpx"
	"github.com/puppe1990/amarra-cais/pkg/cais/session"
	"github.com/puppe1990/amarra-cais/pkg/cais/validate"
)

func (h *AuthHandler) Settings(w http.ResponseWriter, r *http.Request) {
	if _, ok := session.UserID(r); !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	h.renderAuth(w, r, "settings", map[string]any{
		"Title": h.t(r, "settings.title"),
	}, 0)
}

func (h *AuthHandler) SettingsPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := session.UserID(r)
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if err := httpx.ParseFormOrJSON(r); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	current := r.FormValue("current_password")
	password := r.FormValue("password")
	confirm := r.FormValue("password_confirmation")

	user, err := h.store.FindUser(userID)
	if err != nil {
		httpx.ServerError(w, err, h.cfg)
		return
	}

	var errs validate.FieldErrors
	if !session.VerifyPassword(user.PasswordHash, current) {
		errs.Add("current_password", h.t(r, "settings.current_password_invalid"))
	}
	if err := validate.MinLength(password, 8); err != nil {
		errs.Add("password", h.t(r, "auth.password_too_short"))
	}
	if password != confirm {
		errs.Add("password_confirmation", h.t(r, "auth.password_mismatch"))
	}
	if errs.Any() {
		h.renderAuth(w, r, "settings", map[string]any{
			"Title":  h.t(r, "settings.title"),
			"Errors": errs,
		}, http.StatusUnprocessableEntity)
		return
	}

	hash, err := session.HashPassword(password)
	if err != nil {
		httpx.ServerError(w, err, h.cfg)
		return
	}
	if err := h.store.ChangePassword(userID, hash); err != nil {
		httpx.ServerError(w, err, h.cfg)
		return
	}
	if err := session.SignOut(w, h.sessions, r, session.CookieOptionsFromConfig(h.cfg)); err != nil {
		httpx.ServerError(w, err, h.cfg)
		return
	}
	flash.Set(w, "notice", h.t(r, "auth.reset_success"), h.cfg.CookieSecure())
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
`
