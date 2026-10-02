package cais

import (
	"strings"
	"testing"
)

func TestConfig_Validate_rejectsInvalidMaxBodyBytes(t *testing.T) {
	cases := []struct {
		name  string
		value string
	}{
		{name: "text", value: "abc"},
		{name: "zero", value: "0"},
		{name: "negative", value: "-1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("MAX_BODY_BYTES", tc.value)
			cfg := Load()
			err := cfg.Validate()
			if err == nil {
				t.Fatal("expected Validate() error")
			}
			msg := err.Error()
			if !strings.Contains(msg, "MAX_BODY_BYTES") {
				t.Errorf("error %q does not cite MAX_BODY_BYTES", msg)
			}
			if !strings.Contains(msg, tc.value) {
				t.Errorf("error %q does not cite value %q", msg, tc.value)
			}
			if cfg.BodyLimit() != DefaultMaxBodyBytes {
				t.Errorf("BodyLimit() = %d, want default %d after invalid env", cfg.BodyLimit(), DefaultMaxBodyBytes)
			}
		})
	}
}

func TestConfig_Validate_missingMaxBodyBytesUsesDefault(t *testing.T) {
	t.Setenv("MAX_BODY_BYTES", "")
	cfg := Load()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unset MAX_BODY_BYTES should Validate: %v", err)
	}
	if cfg.BodyLimit() != DefaultMaxBodyBytes {
		t.Errorf("BodyLimit() = %d, want %d", cfg.BodyLimit(), DefaultMaxBodyBytes)
	}
}

func TestConfig_Validate_validMaxBodyBytes(t *testing.T) {
	t.Setenv("MAX_BODY_BYTES", "1048576")
	cfg := Load()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("valid MAX_BODY_BYTES should Validate: %v", err)
	}
	if cfg.BodyLimit() != 1<<20 {
		t.Errorf("BodyLimit() = %d, want %d", cfg.BodyLimit(), 1<<20)
	}
}

func TestConfig_Validate_rejectsInvalidEnv(t *testing.T) {
	cases := []struct {
		name    string
		env     map[string]string
		wantVar string
		wantVal string
	}{
		{
			name:    "PORT",
			env:     map[string]string{"PORT": "not-a-port"},
			wantVar: "PORT",
			wantVal: "not-a-port",
		},
		{
			name:    "APP_URL",
			env:     map[string]string{"APP_URL": "not-a-url"},
			wantVar: "APP_URL",
			wantVal: "not-a-url",
		},
		{
			name:    "TRUSTED_PROXIES",
			env:     map[string]string{"TRUSTED_PROXIES": "127.0.0.1, not-an-ip"},
			wantVar: "TRUSTED_PROXIES",
			wantVal: "not-an-ip",
		},
		{
			name:    "CSP_SCRIPT_SRC line break",
			env:     map[string]string{"CSP_SCRIPT_SRC": "https://cdn.example.com\nhttps://evil.example"},
			wantVar: "CSP_SCRIPT_SRC",
			wantVal: "https://cdn.example.com",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("PORT", "")
			t.Setenv("APP_URL", "")
			t.Setenv("TRUSTED_PROXIES", "")
			t.Setenv("CSP_SCRIPT_SRC", "")
			t.Setenv("MAX_BODY_BYTES", "")
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			err := Load().Validate()
			if err == nil {
				t.Fatal("expected Validate() error")
			}
			msg := err.Error()
			if !strings.Contains(msg, tc.wantVar) {
				t.Errorf("error %q does not cite %s", msg, tc.wantVar)
			}
			if !strings.Contains(msg, tc.wantVal) {
				t.Errorf("error %q does not cite %q", msg, tc.wantVal)
			}
		})
	}
}

func TestConfig_Load_barePORTBecomesListenAddr(t *testing.T) {
	t.Setenv("PORT", "4096")
	t.Setenv("ENV", "")
	cfg := Load()
	if cfg.Port != ":4096" {
		t.Errorf("Port = %q, want :4096", cfg.Port)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() after PORT=4096: %v", err)
	}
}

func TestConfig_Validate_requiresPORTInProduction(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("ENV", "production")
	t.Setenv("ADMIN_TOKEN", "secret")
	t.Setenv("APP_URL", "https://example.com")
	cfg := Load()
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() error without PORT in production")
	}
	if !strings.Contains(err.Error(), "PORT") {
		t.Errorf("error %q does not cite PORT", err.Error())
	}
}

func TestConfig_Validate_productionBarePORT(t *testing.T) {
	t.Setenv("PORT", "4096")
	t.Setenv("ENV", "production")
	t.Setenv("ADMIN_TOKEN", "secret")
	t.Setenv("APP_URL", "https://example.com")
	cfg := Load()
	if cfg.Port != ":4096" {
		t.Errorf("Port = %q, want :4096", cfg.Port)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() with PORT=4096 in production: %v", err)
	}
}

func TestConfig_Validate_emptyTrustedProxiesTokensAreUnset(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "  ,  ")
	cfg := Load()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("blank TRUSTED_PROXIES tokens should Validate: %v", err)
	}
	if len(cfg.TrustedProxies) != 0 {
		t.Errorf("TrustedProxies = %v, want empty", cfg.TrustedProxies)
	}
}
