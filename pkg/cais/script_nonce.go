package cais

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
)

type scriptNonceKey struct{}

// NewScriptNonce returns a URL-safe per-request CSP nonce.
func NewScriptNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("csp nonce: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// WithScriptNonce stores the CSP nonce on the request context so view.Write
// and localhost dashboards can stamp matching nonce attributes (#263).
func WithScriptNonce(ctx context.Context, nonce string) context.Context {
	return context.WithValue(ctx, scriptNonceKey{}, nonce)
}

// ScriptNonceFromRequest returns the nonce SecurityHeaders issued, or "".
func ScriptNonceFromRequest(r *http.Request) string {
	if r == nil {
		return ""
	}
	nonce, _ := r.Context().Value(scriptNonceKey{}).(string)
	return nonce
}
