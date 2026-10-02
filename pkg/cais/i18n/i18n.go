package i18n

import (
	"fmt"
	"html/template"
	"strings"
)

const DefaultLocale = "en"

var locales = map[string]map[string]string{
	"en": enMessages,
	"pt": ptMessages,
}

// Catalog holds translated strings for a locale.
type Catalog struct {
	locale   string
	messages map[string]string
}

// DefaultCatalog returns the English catalog.
func DefaultCatalog() *Catalog {
	return NewCatalog(DefaultLocale)
}

// NewCatalog builds a catalog for the given locale tag.
func NewCatalog(locale string) *Catalog {
	return NewCatalogFrom(locale, locales)
}

// NewCatalogFrom builds a catalog from custom locale maps (for app-level translations).
func NewCatalogFrom(locale string, locales map[string]map[string]string) *Catalog {
	tag := normalizeLocale(locale)
	msgs, ok := locales[tag]
	if !ok {
		tag = DefaultLocale
		msgs, ok = locales[tag]
		if !ok {
			return &Catalog{locale: DefaultLocale, messages: map[string]string{}}
		}
	}
	copied := make(map[string]string, len(msgs))
	for k, v := range msgs {
		copied[k] = v
	}
	return &Catalog{locale: tag, messages: copied}
}

// NormalizeLocale maps BCP 47 / underscore tags to catalog keys (pt-BR → pt).
func NormalizeLocale(tag string) string {
	return normalizeLocale(tag)
}

func normalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	locale = strings.ReplaceAll(locale, "-", "_")
	switch {
	case locale == "" || strings.HasPrefix(locale, "en"):
		return "en"
	case strings.HasPrefix(locale, "pt"):
		return "pt"
	case strings.HasPrefix(locale, "es"):
		return "es"
	case strings.HasPrefix(locale, "zh"):
		return "zh"
	default:
		return DefaultLocale
	}
}

// T returns the translation for key, optionally formatting with args.
func (c *Catalog) T(key string, args ...any) string {
	msg, ok := c.messages[key]
	if !ok {
		return key
	}
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}

// Locale returns the normalized locale tag (en, pt).
func (c *Catalog) Locale() string {
	return c.locale
}

// HTMLLang returns the BCP 47 language tag for <html lang>.
func (c *Catalog) HTMLLang() string {
	switch c.locale {
	case "pt":
		return "pt-BR"
	case "es":
		return "es-419"
	case "zh":
		return "zh-CN"
	default:
		return "en"
	}
}

// OGLocale returns the Open Graph locale value.
func (c *Catalog) OGLocale() string {
	switch c.locale {
	case "pt":
		return "pt_BR"
	case "es":
		return "es_419"
	case "zh":
		return "zh_CN"
	default:
		return "en_US"
	}
}

// Funcs returns template helpers: t, htmlLang, ogLocale, localeBase,
// localeTags, localeLabel.
func (c *Catalog) Funcs() template.FuncMap {
	return template.FuncMap{
		"t": func(key string, args ...any) string {
			return c.T(key, args...)
		},
		"htmlLang": func() string { return c.HTMLLang() },
		"ogLocale": func() string { return c.OGLocale() },
		// any: kit current may be missing when .Locale is unset (#209).
		"localeBase": func(tag any) string {
			s, _ := tag.(string)
			return NormalizeLocale(s)
		},
		// localeTags lists kit toggle buttons. A comma string or []string
		// from locales= / page .Locales; empty falls back to en|pt (#300).
		"localeTags":  localeTagsFrom,
		"localeLabel": localeLabelFrom,
	}
}

func localeTagsFrom(v any) []string {
	var raw []string
	switch t := v.(type) {
	case []string:
		raw = t
	case string:
		raw = strings.Split(t, ",")
	}
	out := make([]string, 0, len(raw))
	seen := map[string]bool{}
	for _, part := range raw {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		tag := NormalizeLocale(part)
		if seen[tag] {
			continue
		}
		seen[tag] = true
		out = append(out, tag)
	}
	if len(out) == 0 {
		return []string{"en", "pt"}
	}
	return out
}

func localeLabelFrom(v any) string {
	s, _ := v.(string)
	return strings.ToUpper(NormalizeLocale(s))
}

// MergeFuncs combines i18n funcs with additional template funcs.
func MergeFuncs(c *Catalog, extra template.FuncMap) template.FuncMap {
	out := template.FuncMap{}
	for k, v := range extra {
		out[k] = v
	}
	for k, v := range c.Funcs() {
		out[k] = v
	}
	return out
}
