package passwordreset

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"
)

// DefaultTTL is how long a reset token stays valid.
const DefaultTTL = time.Hour

// Notifier delivers password reset instructions (email in production, log in development).
type Notifier interface {
	NotifyReset(email, token string) error
}

// LogNotifier logs reset links server-side until SMTP is wired.
type LogNotifier struct {
	Out    io.Writer
	AppURL string
}

func (n LogNotifier) NotifyReset(email, token string) error {
	out := n.Out
	if out == nil {
		out = os.Stdout
	}
	link := BuildResetURL(n.AppURL, token)
	_, err := fmt.Fprintf(out, "password reset for %s: %s\n", email, link)
	return err
}

// BuildResetURL returns the absolute reset link for a token.
func BuildResetURL(appURL, token string) string {
	base := strings.TrimRight(appURL, "/")
	if base == "" {
		base = "http://localhost:8080"
	}
	return base + "/reset-password?token=" + url.QueryEscape(token)
}

// NewToken returns a cryptographically random hex token.
func NewToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// Hash returns the hex-encoded SHA-256 digest persisted at rest. Reset tokens
// already carry 256 bits of entropy, so a fast digest is enough: a leaked DB,
// backup or SQL log then holds a non-replayable value instead of a live bearer
// credential, without paying bcrypt on every lookup (#222).
func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
