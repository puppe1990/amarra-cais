package meta

import (
	"strings"
	"testing"
)

func TestAbsoluteURL(t *testing.T) {
	tests := []struct {
		base, path, want string
	}{
		{"", "/static/og.png", "/static/og.png"},
		{"https://app.example.com", "/static/og.png", "https://app.example.com/static/og.png"},
		{"https://app.example.com/", "static/og.png", "https://app.example.com/static/og.png"},
		{"https://app.example.com", "", "https://app.example.com/"},
	}
	for _, tc := range tests {
		if got := AbsoluteURL(tc.base, tc.path); got != tc.want {
			t.Errorf("AbsoluteURL(%q, %q) = %q, want %q", tc.base, tc.path, got, tc.want)
		}
	}
}

func TestDefaultPreview(t *testing.T) {
	p := DefaultPreview("Harbor")
	if p.Title != "Harbor" {
		t.Errorf("Title = %q", p.Title)
	}
	if p.Image != DefaultImagePath {
		t.Errorf("Image = %q", p.Image)
	}
	if p.Locale != "en_US" {
		t.Errorf("Locale = %q", p.Locale)
	}
}

func TestPreviewHTML(t *testing.T) {
	html := PreviewHTML(Preview{
		Title:       "Harbor",
		Description: "A small Go app.",
		SiteName:    "Harbor",
		SiteURL:     "https://app.example.com",
		Path:        "/login",
		Image:       DefaultImagePath,
	})

	for _, want := range []string{
		`property="og:type" content="website"`,
		`property="og:title" content="Harbor"`,
		`property="og:description" content="A small Go app."`,
		`property="og:image" content="https://app.example.com/static/og.png"`,
		`property="og:url" content="https://app.example.com/login"`,
		`name="twitter:card" content="summary_large_image"`,
		`name="twitter:image" content="https://app.example.com/static/og.png"`,
		`rel="canonical" href="https://app.example.com/login"`,
		`property="og:image:alt" content="Harbor"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("PreviewHTML missing %q in:\n%s", want, html)
		}
	}
}

func TestPreviewHTML_imageDimensionsAndAlt(t *testing.T) {
	html := PreviewHTML(Preview{
		Title:       "Harbor",
		Description: "A small Go app.",
		SiteName:    "Harbor",
		SiteURL:     "https://app.example.com",
		Path:        "/",
		Image:       DefaultImagePath,
		ImageWidth:  1200,
		ImageHeight: 630,
		ImageAlt:    "Harbor at dusk",
	})
	for _, want := range []string{
		`property="og:image:width" content="1200"`,
		`property="og:image:height" content="630"`,
		`property="og:image:alt" content="Harbor at dusk"`,
		`rel="canonical" href="https://app.example.com/"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("PreviewHTML missing %q in:\n%s", want, html)
		}
	}
}

func TestSiteFrom(t *testing.T) {
	site := SiteFrom("My App", "https://example.com/")
	if site.AppName != "My App" {
		t.Errorf("AppName = %q", site.AppName)
	}
	if site.AppURL != "https://example.com" {
		t.Errorf("AppURL = %q", site.AppURL)
	}
}

func TestTemplateFuncs_absURL(t *testing.T) {
	fn, ok := TemplateFuncs()["absURL"].(func(string, string) string)
	if !ok {
		t.Fatal("absURL func missing")
	}
	if got := fn("https://example.com", "/static/og.png"); got != "https://example.com/static/og.png" {
		t.Errorf("absURL = %q", got)
	}
}
