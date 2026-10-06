package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeInlineFixture(t *testing.T, dir, html string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "web/templates/pages"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "web/templates/pages/home.html"), []byte(html), 0o644); err != nil {
		t.Fatal(err)
	}
}

// #319: inline script sem nonce e handler inline quebram a policy de produção
// ('self' + nonce) sem o doctor avisar.
func TestCheckInlineScripts_warnsOnScriptWithoutNonce(t *testing.T) {
	dir := t.TempDir()
	writeInlineFixture(t, dir, `<script>console.log(1)</script><script src="/static/js/app.js"></script>`)
	c := checkInlineScripts(dir)
	if c.OK {
		t.Fatal("script inline sem nonce deve avisar")
	}
	if !strings.Contains(c.FixHint, "CSPNonce") {
		t.Errorf("FixHint = %q", c.FixHint)
	}
}

func TestCheckInlineScripts_okWhenNoncePresent(t *testing.T) {
	dir := t.TempDir()
	writeInlineFixture(t, dir, `<script nonce="{{ .CSPNonce }}">console.log(1)</script>`)
	if c := checkInlineScripts(dir); !c.OK {
		t.Fatalf("script com nonce deve passar, got %+v", c)
	}
}

func TestCheckInlineScripts_warnsOnInlineHandler(t *testing.T) {
	dir := t.TempDir()
	writeInlineFixture(t, dir, `<button onclick="window.print()">Print</button>`)
	if c := checkInlineScripts(dir); c.OK {
		t.Fatal("handler inline deve avisar")
	}
}

func TestCheckInlineScripts_warnsOnSwRegisterWithoutScope(t *testing.T) {
	dir := t.TempDir()
	writeInlineFixture(t, dir, `<script nonce="{{ .CSPNonce }}">navigator.serviceWorker.register("/static/js/sw.js")</script>`)
	c := checkInlineScripts(dir)
	if c.OK {
		t.Fatal("register sem scope deve avisar")
	}
	if !strings.Contains(c.Detail+c.FixHint, "scope") {
		t.Errorf("want scope hint, got %+v", c)
	}
}
