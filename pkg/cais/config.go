package cais

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/puppe1990/amarra-cais/pkg/cais/dotenv"
)

type Config struct {
	Port              string
	DBPath            string
	Env               string
	AppURL            string
	AdminToken        string
	Locale            string
	LogFormat         string
	StaticDir         string
	TemplatesDir      string
	TrustedProxies    []string
	PermissionsPolicy string
	CSPStyleSrc       string
	CSPConnectSrc     string
	CSPMediaSrc       string
	CSPImgSrc         string
	CSPFontSrc        string
	CSPScriptSrc      string
	// MaxBodyBytes caps the total size of a request body before any parse, so
	// oversized uploads are rejected before net/http spills them to temp files
	// (#221). Zero falls back to DefaultMaxBodyBytes; raise via MAX_BODY_BYTES.
	MaxBodyBytes int64
	// envFaults records explicit env values Load could not apply (#286).
	// Validate surfaces them; a missing variable is not a fault.
	envFaults []string
}

// DefaultMaxBodyBytes is the safe total request-body cap: ParseMultipartForm's
// 32 MiB only bounds the in-memory fraction and never the request size (#221).
const DefaultMaxBodyBytes int64 = 32 << 20

// BodyLimit returns the configured total request-body cap, falling back to the
// safe default when unset/invalid.
func (c Config) BodyLimit() int64 {
	if c.MaxBodyBytes > 0 {
		return c.MaxBodyBytes
	}
	return DefaultMaxBodyBytes
}

func Load() Config {
	// Scaffold .env.example is otherwise dead: doctor parses .env, Load did not (#153).
	_ = dotenv.LoadFile(".env")

	cfg := Config{
		Port:   ":8080",
		DBPath: "./data/app.db",
		Env:    "development",
		Locale: "en",
	}

	if v := os.Getenv("PORT"); v != "" {
		cfg.Port = v
		if _, _, err := parseListenPort(strings.TrimSpace(v)); err != nil {
			cfg.noteInvalidEnv("PORT", v, "not a listen address; use :8080 or 127.0.0.1:8080")
		}
	}
	if v := os.Getenv("DB_PATH"); v != "" {
		cfg.DBPath = v
	}
	if v := os.Getenv("ENV"); v != "" {
		cfg.Env = v
	}
	if v := os.Getenv("APP_URL"); v != "" {
		cfg.AppURL = v
		if !validAbsoluteHTTPURL(v) {
			cfg.noteInvalidEnv("APP_URL", v, "not an absolute http(s) URL; example: https://app.example.com")
		}
	}
	if v := os.Getenv("ADMIN_TOKEN"); v != "" {
		cfg.AdminToken = v
	}
	if v := os.Getenv("LOCALE"); v != "" {
		cfg.Locale = v
	}
	if v := os.Getenv("LOG_FORMAT"); v != "" {
		cfg.LogFormat = v
	}
	if v := os.Getenv("STATIC_DIR"); v != "" {
		cfg.StaticDir = v
	}
	if v := os.Getenv("TEMPLATES_DIR"); v != "" {
		cfg.TemplatesDir = v
	}
	if v := os.Getenv("TRUSTED_PROXIES"); v != "" {
		for _, ip := range strings.Split(v, ",") {
			if ip = strings.TrimSpace(ip); ip == "" {
				continue
			}
			if !validProxyEntry(ip) {
				cfg.noteInvalidEnv("TRUSTED_PROXIES", ip, "not an IP or CIDR; use values like 127.0.0.1 or 10.0.0.0/8")
				continue
			}
			cfg.TrustedProxies = append(cfg.TrustedProxies, ip)
		}
	}
	if v := os.Getenv("PERMISSIONS_POLICY"); v != "" {
		cfg.PermissionsPolicy = v
	} else {
		// #262: camera=() in every env. Apps that scan barcodes set
		// PERMISSIONS_POLICY=camera=(self), microphone=(), geolocation=().
		cfg.PermissionsPolicy = "camera=(), microphone=(), geolocation=()"
	}
	cfg.setCSPFromEnv("CSP_STYLE_SRC", &cfg.CSPStyleSrc)
	cfg.setCSPFromEnv("CSP_CONNECT_SRC", &cfg.CSPConnectSrc)
	if v := os.Getenv("CSP_MEDIA_SRC"); v != "" {
		cfg.CSPMediaSrc = v
		cfg.rejectHeaderBreak("CSP_MEDIA_SRC", v)
	} else if cfg.Env == "development" {
		cfg.CSPMediaSrc = "blob:"
	}
	cfg.setCSPFromEnv("CSP_IMG_SRC", &cfg.CSPImgSrc)
	cfg.setCSPFromEnv("CSP_FONT_SRC", &cfg.CSPFontSrc)
	cfg.setCSPFromEnv("CSP_SCRIPT_SRC", &cfg.CSPScriptSrc)
	cfg.setMaxBodyBytesFromEnv(os.Getenv("MAX_BODY_BYTES"))

	return cfg
}

func (c Config) CookieSecure() bool {
	return c.Env == "production"
}

func (c Config) SanitizeErrors() bool {
	return c.Env == "production"
}

// LogJSON reports whether request/SQL logs should emit structured JSON lines.
// Default: JSON in development and production; LOG_FORMAT=text opts out; LOG_FORMAT=json forces JSON.
func (c Config) LogJSON() bool {
	switch strings.ToLower(strings.TrimSpace(c.LogFormat)) {
	case "json":
		return true
	case "text":
		return false
	default:
		return c.Env == "development" || c.Env == "production"
	}
}

// Validate checks required settings for the active environment and explicit
// env values Load could not apply (#286).
func (c Config) Validate() error {
	if err := c.envFaultsError(); err != nil {
		return err
	}
	if c.Env == "production" && c.AdminToken == "" {
		return fmt.Errorf("ADMIN_TOKEN is required when ENV=production")
	}
	if c.Env == "production" && c.AppURL == "" {
		return fmt.Errorf("APP_URL is required when ENV=production")
	}
	return nil
}

func (c *Config) setMaxBodyBytesFromEnv(raw string) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n <= 0 {
		c.noteInvalidEnv("MAX_BODY_BYTES", raw, "not a positive integer; set a byte count such as 33554432, or unset MAX_BODY_BYTES to use the 32 MiB default")
		return
	}
	c.MaxBodyBytes = n
}

func (c *Config) setCSPFromEnv(name string, dest *string) {
	v := os.Getenv(name)
	if v == "" {
		return
	}
	*dest = v
	c.rejectHeaderBreak(name, v)
}

func (c *Config) rejectHeaderBreak(name, raw string) {
	if strings.ContainsAny(raw, "\r\n") {
		c.noteInvalidEnv(name, raw, "contains a line break; CSP extras must be a single-line value")
	}
}

func (c *Config) noteInvalidEnv(name, value, reason string) {
	c.envFaults = append(c.envFaults, fmt.Sprintf("%s=%q is invalid: %s", name, value, reason))
}

func (c Config) envFaultsError() error {
	if len(c.envFaults) == 0 {
		return nil
	}
	if len(c.envFaults) == 1 {
		return fmt.Errorf("%s", c.envFaults[0])
	}
	return fmt.Errorf("%s", strings.Join(c.envFaults, "; "))
}

func validAbsoluteHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}

func validProxyEntry(s string) bool {
	if strings.Contains(s, "/") {
		_, _, err := net.ParseCIDR(s)
		return err == nil
	}
	return net.ParseIP(s) != nil
}
