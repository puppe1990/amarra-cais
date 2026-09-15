package pwa

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func installedStatic(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := InstallForAmarra(dir, "Demo"); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(dir, "web", "static")
}

func readStatic(t *testing.T, staticDir, rel string) []byte {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(staticDir, rel))
	if err != nil {
		t.Fatal(err)
	}
	return body
}

// A scaffold must not carry the framework's mark: the default icon is a neutral
// placeholder, so an installed PWA or a shared link is not someone else's brand (#64).
func TestWriteStatic_defaultIconIsNeutralPlaceholder(t *testing.T) {
	staticDir := installedStatic(t)
	body := readStatic(t, staticDir, "icons/icon.png")

	if len(body) > 20_000 {
		t.Errorf("icons/icon.png is %d bytes — designed brand art is back (placeholder tiles stay tiny)", len(body))
	}
	img, err := png.Decode(bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	b := img.Bounds()
	corner := img.At(b.Min.X, b.Min.Y)
	for _, pt := range [][2]int{{b.Max.X - 1, b.Min.Y}, {b.Min.X, b.Max.Y - 1}, {b.Max.X - 1, b.Max.Y - 1}} {
		if img.At(pt[0], pt[1]) != corner {
			t.Fatal("placeholder must be a plain tile (uniform background)")
		}
	}
}

// Splitting `any` and `maskable` is what stops Android from cropping the icon:
// one "any maskable" entry declares the same file safe for both (#64).
func TestWriteStatic_manifestSplitsAnyAndMaskable(t *testing.T) {
	staticDir := installedStatic(t)
	manifest := string(readStatic(t, staticDir, "manifest.webmanifest"))

	if strings.Contains(manifest, "any maskable") {
		t.Errorf("manifest must not declare a single any maskable entry:\n%s", manifest)
	}
	if got := strings.Count(manifest, `"purpose": "any"`); got != 2 {
		t.Errorf("manifest should keep 192 + 512 as purpose any, got %d:\n%s", got, manifest)
	}
	if !strings.Contains(manifest, `"purpose": "maskable"`) {
		t.Errorf("manifest missing a maskable entry:\n%s", manifest)
	}
	if !strings.Contains(manifest, "/static/icons/icon-512-maskable.png") {
		t.Errorf("maskable entry should point at its own padded file:\n%s", manifest)
	}
}

// The maskable file keeps the mark inside the safe zone, so replacing
// icon-512.png with a full-bleed logo does not get cropped (#64).
func TestWriteStatic_maskableIconKeepsSafeZone(t *testing.T) {
	staticDir := installedStatic(t)
	body := readStatic(t, staticDir, "icons/icon-512-maskable.png")

	img, err := png.Decode(bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 512 || b.Dy() != 512 {
		t.Fatalf("maskable icon = %dx%d, want 512x512", b.Dx(), b.Dy())
	}
	b := img.Bounds()
	bg := img.At(b.Min.X, b.Min.Y)
	if !cornersAreBackground(img) {
		t.Error("maskable corners must be background (safe zone)")
	}
	if img.At(b.Dx()/2, b.Dy()/2) == bg {
		t.Error("maskable icon lost its mark")
	}
}

func cornersAreBackground(img image.Image) bool {
	b := img.Bounds()
	bg := img.At(b.Min.X, b.Min.Y)
	for _, pt := range [][2]int{{b.Max.X - 1, b.Min.Y}, {b.Min.X, b.Max.Y - 1}, {b.Max.X - 1, b.Max.Y - 1}} {
		if img.At(pt[0], pt[1]) != bg {
			return false
		}
	}
	return true
}
