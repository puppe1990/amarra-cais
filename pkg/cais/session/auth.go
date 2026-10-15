package session

import (
	"fmt"
	"net/http"
)

// SignIn rotates the session: deletes any existing token before issuing a new one.
// Prevents reuse of a pre-login token after authentication succeeds.
func SignIn(w http.ResponseWriter, store Store, r *http.Request, userID int64, opts CookieOptions) error {
	if token := TokenFromRequest(r); token != "" {
		if err := store.Delete(token); err != nil {
			return fmt.Errorf("revoke session: %w", err)
		}
	}
	newToken, err := store.Create(userID)
	if err != nil {
		return err
	}
	SetCookie(w, newToken, opts)
	return nil
}

// SignOut clears the cookie only after server-side revocation succeeds (#352).
func SignOut(w http.ResponseWriter, store Store, r *http.Request, opts CookieOptions) error {
	if token := TokenFromRequest(r); token != "" {
		if err := store.Delete(token); err != nil {
			return fmt.Errorf("revoke session: %w", err)
		}
	}
	ClearCookie(w, opts)
	return nil
}
