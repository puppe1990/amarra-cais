package netutil

import (
	"strings"
	"testing"
)

func TestHealthPayload_includesStatusAndLANURLs(t *testing.T) {
	payload := HealthPayload("ok", ":8080", "development")
	if payload["status"] != "ok" {
		t.Errorf("status = %v", payload["status"])
	}
	urls, ok := payload["lan_urls"].([]string)
	if !ok {
		t.Fatalf("lan_urls type = %T", payload["lan_urls"])
	}
	for _, u := range urls {
		if !strings.HasPrefix(u, "http://") {
			t.Errorf("LAN URL %q should start with http://", u)
		}
		if strings.Contains(u, "http://http://") {
			t.Errorf("malformed double scheme in %q", u)
		}
	}
}

// #131: /health is public; lan_urls expose the host's RFC1918 topology and
// must not ship in production.
func TestHealthPayload_omitsLANURLsInProduction(t *testing.T) {
	payload := HealthPayload("ok", ":8080", "production")
	if _, ok := payload["lan_urls"]; ok {
		t.Fatalf("production payload leaked lan_urls: %v", payload["lan_urls"])
	}
	if payload["status"] != "ok" {
		t.Fatalf("status = %v", payload["status"])
	}
}
