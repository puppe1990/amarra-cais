// Brand tiles, maskable icons, and OG image writes (#288).
package pwa

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

// writeOGImage copies the neutral 1200x630 preview shipped with the framework.
// Apps replace web/static/og.png with their own art (#64).
func writeOGImage(path string, force bool) error {
	if !force && fileExists(path) {
		return nil
	}
	return copyAsset("assets/og.png", path)
}

func fill(img *image.RGBA, c color.RGBA) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			img.Set(x, y, c)
		}
	}
}

func encodePNGBytes(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// iconSource returns the icon an app scaffolds with: the caller's brand file
// when configured, otherwise the framework's neutral placeholder tile.
func iconSource(iconPath string) ([]byte, error) {
	if iconPath != "" {
		return os.ReadFile(iconPath)
	}
	return assets.ReadFile("assets/icon.png")
}

func writeAppIcons(dir string, iconPath string, force bool) error {
	if !force && iconsComplete(dir) {
		return nil
	}
	data, err := iconSource(iconPath)
	if err != nil {
		return err
	}
	src, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return err
	}
	if err := writeIcon(dir, "icon.png", data, force); err != nil {
		return err
	}
	for _, size := range []int{192, 512} {
		dst := resizeNearest(src, size, size)
		body, err := encodePNGBytes(dst)
		if err != nil {
			return err
		}
		if err := writeIcon(dir, fmt.Sprintf("icon-%d.png", size), body, force); err != nil {
			return err
		}
	}
	// Separate maskable file: platforms crop `any` icons, so keep a version whose
	// content sits inside the safe zone (#64).
	body, err := encodePNGBytes(maskableIcon(src, 512))
	if err != nil {
		return err
	}
	return writeIcon(dir, "icon-512-maskable.png", body, force)
}

// writeIcon leaves an existing icon untouched unless force, so a refresh can
// backfill a missing file (e.g. maskable) without clobbering replaced icons (#186).
func writeIcon(dir, name string, data []byte, force bool) error {
	path := filepath.Join(dir, name)
	if !force && fileExists(path) {
		return nil
	}
	return writeFileSafe(path, data)
}

func iconsComplete(dir string) bool {
	for _, name := range []string{"icon.png", "icon-192.png", "icon-512.png", "icon-512-maskable.png"} {
		if !fileExists(filepath.Join(dir, name)) {
			return false
		}
	}
	return true
}

// maskableIcon insets src to 80% of the canvas over its own corner color, the
// Android safe zone for maskable icons.
func maskableIcon(src image.Image, size int) *image.RGBA {
	inner := resizeNearest(src, size*4/5, size*4/5)
	bounds := src.Bounds()
	bg := color.RGBAModel.Convert(src.At(bounds.Min.X, bounds.Min.Y)).(color.RGBA)
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	fill(dst, bg)
	offset := (size - inner.Bounds().Dx()) / 2
	for y := 0; y < inner.Bounds().Dy(); y++ {
		for x := 0; x < inner.Bounds().Dx(); x++ {
			dst.Set(offset+x, offset+y, inner.At(x, y))
		}
	}
	return dst
}

// DefaultBrandAssets returns the brand files a scaffold writes, keyed by their
// path under web/static. doctor compares app files against them to warn while
// the app still ships the placeholder (#64).
func DefaultBrandAssets() map[string][]byte {
	out := map[string][]byte{}
	data, err := assets.ReadFile("assets/icon.png")
	if err != nil {
		return out
	}
	src, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return out
	}
	out["icons/icon.png"] = data
	encoded := map[string]image.Image{
		"icons/icon-192.png":          resizeNearest(src, 192, 192),
		"icons/icon-512.png":          resizeNearest(src, 512, 512),
		"icons/icon-512-maskable.png": maskableIcon(src, 512),
	}
	for rel, img := range encoded {
		body, err := encodePNGBytes(img)
		if err != nil {
			return map[string][]byte{}
		}
		out[rel] = body
	}
	if og, err := assets.ReadFile("assets/og.png"); err == nil {
		out["og.png"] = og
	}
	if fav, err := assets.ReadFile("assets/favicon.svg"); err == nil {
		out["favicon.svg"] = fav
	}
	return out
}

func resizeNearest(src image.Image, w, h int) *image.RGBA {
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	sw, sh := bounds.Dx(), bounds.Dy()
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			sx := bounds.Min.X + x*sw/w
			sy := bounds.Min.Y + y*sh/h
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}
